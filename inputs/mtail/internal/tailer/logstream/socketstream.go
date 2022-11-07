// Copyright 2020 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package logstream

import (
	"bytes"
	"context"
	"log"
	"net"
	"sync"
	"time"

	"flashcat.cloud/categraf/inputs/mtail/internal/logline"
	"flashcat.cloud/categraf/inputs/mtail/internal/waker"
)

type socketStream struct {
	ctx   context.Context
	lines chan<- *logline.LogLine

	oneShot bool
	scheme  string // URL Scheme to listen with, either tcp or unix
	address string // Given name for the underlying socket path on the filesystem or host/port.

	mu           sync.RWMutex // protects following fields
	completed    bool         // This socketStream is completed and can no longer be used.
	lastReadTime time.Time    // Last time a log line was read from this socket

	stopOnce sync.Once     // Ensure stopChan only closed once.
	stopChan chan struct{} // Close to start graceful shutdown.
}

func newSocketStream(ctx context.Context, wg *sync.WaitGroup, waker waker.Waker, scheme, address string, lines chan<- *logline.LogLine, oneShot bool) (LogStream, error) {
	if address == "" {
		return nil, ErrEmptySocketAddress
	}
	ss := &socketStream{ctx: ctx, oneShot: oneShot, scheme: scheme, address: address, lastReadTime: time.Now(), lines: lines, stopChan: make(chan struct{})}
	if err := ss.stream(ctx, wg, waker); err != nil {
		return nil, err
	}
	return ss, nil
}

func (ss *socketStream) LastReadTime() time.Time {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.lastReadTime
}

// stream starts goroutines to read data from the stream socket, until Stop is called or the context is cancelled.
func (ss *socketStream) stream(ctx context.Context, wg *sync.WaitGroup, waker waker.Waker) error {
	l, err := net.Listen(ss.scheme, ss.address)
	if err != nil {
		logErrors.Add(ss.address, 1)
		return err
	}
	// glog.V(2).Infof("opened new socket listener %v", l)

	initDone := make(chan struct{})
	// Set up for shutdown
	wg.Add(1)
	go func() {
		defer wg.Done()
		// If oneshot, wait only for the one conn handler to start, otherwise wait for context Done or stopChan.
		<-initDone
		if !ss.oneShot {
			select {
			case <-ctx.Done():
			case <-ss.stopChan:
			}
		}
		// glog.V(2).Infof("%v: closing listener", l)
		err := l.Close()
		if err != nil {
			log.Println(err)
		}
		ss.mu.Lock()
		ss.completed = true
		ss.mu.Unlock()
	}()

	acceptConn := func() error {
		c, err := l.Accept()
		if err != nil {
			log.Println(err)
			return err
		}
		// glog.V(2).Infof("%v: got new conn %v", l, c)
		wg.Add(1)
		go ss.handleConn(ctx, wg, waker, c)
		return nil
	}

	if ss.oneShot {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := acceptConn(); err != nil {
				log.Println(err)
			}
			log.Println("oneshot mode, retuning")
			close(initDone)
		}()
		return nil
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := acceptConn(); err != nil {
				return
			}
		}
	}()
	close(initDone)
	return nil
}

func (ss *socketStream) handleConn(ctx context.Context, wg *sync.WaitGroup, waker waker.Waker, c net.Conn) {
	defer wg.Done()
	b := make([]byte, defaultReadBufferSize)
	partial := bytes.NewBufferString("")
	var total int
	defer func() {
		// glog.V(2).Infof("%v: read total %d bytes from %s", c, total, ss.address)
		// glog.V(2).Infof("%v: closing connection", c)
		err := c.Close()
		if err != nil {
			logErrors.Add(ss.address, 1)
			log.Println(err)
		}
		logCloses.Add(ss.address, 1)
	}()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	SetReadDeadlineOnDone(ctx, c)

	for {
		n, err := c.Read(b)
		// glog.V(2).Infof("%v: read %d bytes, err is %v", c, n, err)

		if n > 0 {
			total += n
			decodeAndSend(ss.ctx, ss.lines, ss.address, n, b[:n], partial)
			ss.mu.Lock()
			ss.lastReadTime = time.Now()
			ss.mu.Unlock()
		}

		if err != nil && IsEndOrCancel(err) {
			if partial.Len() > 0 {
				sendLine(ctx, ss.address, partial, ss.lines)
			}
			// glog.V(2).Infof("%v: exiting, conn has error %s", c, err)

			return
		}

		// Yield and wait
		// glog.V(2).Infof("%v: waiting", c)
		select {
		case <-ctx.Done():
			// Exit immediately; cancelled context will cause the next read to be interrupted and exit anyway, so no point waiting to loop.
			return
		case <-ss.stopChan:
			// Stop after connection is closed.
			// glog.V(2).Infof("%v: stopchan closed, exiting after next read timeout", c)
		case <-waker.Wake():
			// sleep until next Wake()
			// glog.V(2).Infof("%v: Wake received", c)
		}
	}
}

func (ss *socketStream) IsComplete() bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.completed
}

// Stop implements the Logstream interface.
// Stop will close the listener so no new connections will be accepted, and close all current connections once they have been closed by their peers.
func (ss *socketStream) Stop() {
	ss.stopOnce.Do(func() {
		log.Println("signalling stop at next EOF")
		close(ss.stopChan)
	})
}
