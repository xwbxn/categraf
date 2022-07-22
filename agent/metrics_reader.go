package agent

import (
	"log"
	"strings"
	"time"

	"flashcat.cloud/categraf/config"
	"flashcat.cloud/categraf/house"
	"flashcat.cloud/categraf/inputs"
	"flashcat.cloud/categraf/pkg/runtimex"
	"flashcat.cloud/categraf/types"
	"flashcat.cloud/categraf/writer"
	"github.com/toolkits/pkg/container/list"
)

const agentHostnameLabelKey = "agent_hostname"

var metricReplacer = strings.NewReplacer("-", "_", ".", "_", " ", "_", "'", "_", "\"", "_")

type InputReader struct {
	inputName string
	input     inputs.Input
	quitChan  chan struct{}
	queue     chan *types.Sample
}

func NewInputReader(inputName string, in inputs.Input) *InputReader {
	return &InputReader{
		inputName: inputName,
		input:     in,
		quitChan:  make(chan struct{}, 1),
		queue:     make(chan *types.Sample, config.Config.WriterOpt.ChanSize),
	}
}

func (r *InputReader) Start() {
	// start consumer goroutines
	go r.read()

	// start collector instance
	go r.startInput()
}

func (r *InputReader) Stop() {
	r.quitChan <- struct{}{}
	r.input.Drop()
}

func (r *InputReader) startInput() {
	interval := config.GetInterval()
	if r.input.GetInterval() > 0 {
		interval = time.Duration(r.input.GetInterval())
	}

	for {
		select {
		case <-r.quitChan:
			close(r.quitChan)
			close(r.queue)
			return
		default:
			time.Sleep(interval)
			var start time.Time
			if config.Config.DebugMode {
				start = time.Now()
				log.Println("D!", r.inputName, ": before gather once")
			}

			r.gatherOnce()

			if config.Config.DebugMode {
				ms := time.Since(start).Milliseconds()
				log.Println("D!", r.inputName, ": after gather once,", "duration:", ms, "ms")
			}
		}
	}
}

func (r *InputReader) gatherOnce() {
	defer func() {
		if rc := recover(); rc != nil {
			log.Println("E!", r.inputName, ": gather metrics panic:", r, string(runtimex.Stack(3)))
		}
	}()

	// gather
	slist := list.NewSafeList()
	r.input.Gather(slist)

	// handle result
	samples := slist.PopBackAll()

	size := len(samples)
	if size == 0 {
		return
	}

	if config.Config.DebugMode {
		log.Println("D!", r.inputName, ": gathered samples size:", size)
	}

	now := time.Now()
	for i := 0; i < size; i++ {
		if samples[i] == nil {
			continue
		}

		s := samples[i].(*types.Sample)
		if s == nil {
			continue
		}

		if s.Timestamp.IsZero() {
			s.Timestamp = now
		}

		if len(r.input.Prefix()) > 0 {
			s.Metric = r.input.Prefix() + "_" + metricReplacer.Replace(s.Metric)
		} else {
			s.Metric = metricReplacer.Replace(s.Metric)
		}

		if s.Labels == nil {
			s.Labels = make(map[string]string)
		}

		// add label: agent_hostname
		if _, has := s.Labels[agentHostnameLabelKey]; !has {
			if !config.Config.Global.OmitHostname {
				s.Labels[agentHostnameLabelKey] = config.Config.GetHostname()
			}
		}

		// add global labels
		for k, v := range config.Config.Global.Labels {
			if _, has := s.Labels[k]; !has {
				s.Labels[k] = v
			}
		}

		// write to remote write queue
		r.queue <- s

		// write to clickhouse queue
		house.MetricsHouse.Push(s)
	}
}

func (r *InputReader) read() {
	batch := config.Config.WriterOpt.Batch
	if batch <= 0 {
		batch = 2000
	}

	series := make([]*types.Sample, 0, batch)

	var count int

	for {
		select {
		case item, open := <-r.queue:
			if !open {
				// queue closed
				return
			}

			if item == nil {
				continue
			}

			series = append(series, item)
			count++
			if count >= batch {
				writer.PostSeries(series)
				count = 0
				series = make([]*types.Sample, 0, batch)
			}
		default:
			if len(series) > 0 {
				writer.PostSeries(series)
				count = 0
				series = make([]*types.Sample, 0, batch)
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}
