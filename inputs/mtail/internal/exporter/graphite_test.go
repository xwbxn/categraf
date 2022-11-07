// Copyright 2021 Adam Romanek <romanek.adam@gmail.com>
// This file is available under the Apache license.

package exporter

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"flashcat.cloud/categraf/inputs/mtail/internal/metrics"
	"flashcat.cloud/categraf/inputs/mtail/internal/metrics/datum"
	"flashcat.cloud/categraf/inputs/mtail/internal/testutil"
)

var handleGraphiteTests = []struct {
	name     string
	metrics  []*metrics.Metric
	expected string
}{
	{
		"empty",
		[]*metrics.Metric{},
		"",
	},
	{
		"single",
		[]*metrics.Metric{
			{
				Name:        "foo",
				Program:     "test",
				Kind:        metrics.Counter,
				LabelValues: []*metrics.LabelValue{{Labels: []string{}, Value: datum.MakeInt(1, time.Unix(0, 0))}},
			},
		},
		"foobar.test.foo 1 0\n",
	},
}

func TestHandleGraphite(t *testing.T) {
	*graphitePrefix = "foobar."
	for _, tc := range handleGraphiteTests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			var wg sync.WaitGroup
			ms := metrics.NewStore()
			for _, metric := range tc.metrics {
				testutil.FatalIfErr(t, ms.Add(metric))
			}
			e, err := New(ctx, &wg, ms, Hostname("gunstar"))
			testutil.FatalIfErr(t, err)
			response := httptest.NewRecorder()
			e.HandleGraphite(response, &http.Request{})
			if response.Code != 200 {
				t.Errorf("response code not 200: %d", response.Code)
			}
			b, err := ioutil.ReadAll(response.Body)
			if err != nil {
				t.Errorf("failed to read response %s", err)
			}
			testutil.ExpectNoDiff(t, tc.expected, string(b), testutil.IgnoreUnexported(sync.RWMutex{}))
			cancel()
			wg.Wait()
		})
	}
}
