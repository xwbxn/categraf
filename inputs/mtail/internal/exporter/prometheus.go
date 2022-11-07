// Copyright 2015 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package exporter

import (
	"expvar"
	"fmt"
	"io"
	"log"
	"strings"

	"flashcat.cloud/categraf/inputs/mtail/internal/metrics"
	"flashcat.cloud/categraf/inputs/mtail/internal/metrics/datum"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

var metricExportTotal = expvar.NewInt("metric_export_total")

func noHyphens(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

// Describe implements the prometheus.Collector interface.
func (e *Exporter) Describe(c chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(e, c)
}

// Collect implements the prometheus.Collector interface.
func (e *Exporter) Collect(c chan<- prometheus.Metric) {
	lastMetric := ""
	lastSource := ""

	/* #nosec G104 always retursn nil */
	e.store.Range(func(m *metrics.Metric) error {
		m.RLock()
		// We don't have a way of converting text metrics to prometheus format.
		if m.Kind == metrics.Text {
			m.RUnlock()
			return nil
		}
		metricExportTotal.Add(1)

		lsc := make(chan *metrics.LabelSet)
		go m.EmitLabelSets(lsc)
		for ls := range lsc {
			if lastMetric != m.Name {
				// glog.V(2).Infof("setting source to %s", m.Source)
				lastSource = m.Source
				lastMetric = m.Name
			}
			var keys []string
			var vals []string
			if !e.omitProgLabel {
				keys = append(keys, "prog")
				vals = append(vals, m.Program)
			}
			for k, v := range ls.Labels {
				keys = append(keys, k)
				vals = append(vals, v)
			}
			var pM prometheus.Metric
			var err error
			if m.Kind == metrics.Histogram {
				pM, err = prometheus.NewConstHistogram(
					prometheus.NewDesc(noHyphens(m.Name),
						fmt.Sprintf("defined at %s", lastSource), keys, nil),
					datum.GetBucketsCount(ls.Datum),
					datum.GetBucketsSum(ls.Datum),
					datum.GetBucketsCumByMax(ls.Datum),
					vals...)
			} else {
				pM, err = prometheus.NewConstMetric(
					prometheus.NewDesc(noHyphens(m.Name),
						fmt.Sprintf("defined at %s", lastSource), keys, nil),
					promTypeForKind(m.Kind),
					promValueForDatum(ls.Datum),
					vals...)
			}
			if err != nil {
				log.Println(err)
				return nil
			}
			// By default no timestamp is emitted to Prometheus. Setting a
			// timestamp is not recommended. It can lead to unexpected results
			// if the timestamp is not updated or moved fowarded enough to avoid
			// triggering Promtheus staleness handling.
			// Read more in docs/faq.md
			if e.emitTimestamp {
				c <- prometheus.NewMetricWithTimestamp(ls.Datum.TimeUTC(), pM)
			} else {
				c <- pM
			}
		}
		m.RUnlock()
		return nil
	})
}

// Write is used to write Prometheus metrics to an io.Writer.
func (e *Exporter) Write(w io.Writer) error {
	reg := prometheus.NewRegistry()
	err := reg.Register(e)
	if err != nil {
		return err
	}
	mfs, err := reg.Gather()
	if err != nil {
		return err
	}
	enc := expfmt.NewEncoder(w, expfmt.FmtText)
	for _, mf := range mfs {
		err := enc.Encode(mf)
		if err != nil {
			return err
		}
	}
	return nil
}

func promTypeForKind(k metrics.Kind) prometheus.ValueType {
	switch k {
	case metrics.Counter:
		return prometheus.CounterValue
	case metrics.Gauge:
		return prometheus.GaugeValue
	case metrics.Timer:
		return prometheus.GaugeValue
	}
	return prometheus.UntypedValue
}

func promValueForDatum(d datum.Datum) float64 {
	switch n := d.(type) {
	case *datum.Int:
		return float64(n.Get())
	case *datum.Float:
		return n.Get()
	}
	return 0.
}
