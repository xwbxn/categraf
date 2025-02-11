package upgrade

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTP fetcher uses HEAD requests to poll the status of a given
// file. If it detects this file has been updated, it will fetch
// and return its io.Reader stream.
type HTTP struct {
	//URL to poll for new binaries
	URL           string
	Interval      time.Duration
	CheckHeaders  []string
	BasicAuthUser string
	BasicAuthPass string
	//internal state
	delay bool
	lasts map[string]string
	hmac  string
}

// if any of these change, the binary has been updated
// var defaultHTTPCheckHeaders = []string{"ETag", "If-Modified-Since", "Last-Modified", "Content-Length"}
var defaultHTTPCheckHeaders = []string{"client-version"}

// Init validates the provided config
func (h *HTTP) Init() error {
	//apply defaults
	if h.URL == "" {
		return fmt.Errorf("URL required")
	}
	h.lasts = map[string]string{}
	if h.Interval == 0 {
		h.Interval = 5 * time.Minute
	}
	if h.CheckHeaders == nil {
		h.CheckHeaders = defaultHTTPCheckHeaders
	}
	return nil
}

// Fetch the binary from the provided URL
func (h *HTTP) Fetch() (io.Reader, error) {
	//delay fetches after first
	if h.delay {
		time.Sleep(h.Interval)
	}
	h.delay = true

	//status check using HEAD
	req, err := http.NewRequest("HEAD", h.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("E! failed to new upgrade request: (%s)", err)
	}

	if h.BasicAuthPass != "" {
		req.SetBasicAuth(h.BasicAuthUser, h.BasicAuthPass)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HEAD request failed (%s)", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HEAD request failed (status code %d)", resp.StatusCode)
	}
	//if all headers match, skip update
	matches, total := 0, 0
	for _, header := range h.CheckHeaders {
		if curr := resp.Header.Get(header); curr != "" {
			if last, ok := h.lasts[header]; ok && last == curr {
				matches++
			}
			h.lasts[header] = curr
			total++
		}
	}
	if matches == total {
		return nil, nil //skip, file match
	}
	//binary fetch using GET
	req, err = http.NewRequest("GET", h.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("E! failed to new upgrade request: (%s)", err)
	}

	if h.BasicAuthPass != "" {
		req.SetBasicAuth(h.BasicAuthUser, h.BasicAuthPass)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET request failed (%s)", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET request failed (status code %d)", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") == "application/gzip" {
		return gzip.NewReader(resp.Body)
	}
	//success!
	return resp.Body, nil
}
