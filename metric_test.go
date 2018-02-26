package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestParseInfo(t *testing.T) {
	for k, v := range map[string]map[string]string{
		"type=memory": {
			"type": "memory",
		},
		"type=memory;objects=103;sub-objects=0;master-objects=103;master-sub-objects=0;prole-objects=0;prole-sub-objects=0": {
			"type":               "memory",
			"objects":            "103",
			"sub-objects":        "0",
			"master-objects":     "103",
			"master-sub-objects": "0",
			"prole-objects":      "0",
			"prole-sub-objects":  "0",
		},
		"version=my=version": {
			"version": "my=version",
		},
		"ns=vk:set=user:objects=2501:memory_data_bytes=0:deleting=false": {
			"ns":                "vk",
			"set":               "user",
			"objects":           "2501",
			"memory_data_bytes": "0",
			"deleting":          "false",
		},
	} {
		if have, want := parseInfo(k), v; !reflect.DeepEqual(have, want) {
			t.Errorf("have %+v, want %+v", have, want)
		}
	}
}

func TestInfoCollect(t *testing.T) {
	type cas struct {
		payload string
		field   string
		metric  cmetric
		want    string
	}
	for n, c := range []cas{
		{
			payload: "counter-1=3.14:gauge-1=6.12:flag=false:counter-2=6.66",
			field:   "gauge-1",
			metric: cmetric{
				typ: prometheus.GaugeValue,
				desc: prometheus.NewDesc(
					"g1",
					"My first gauge",
					nil,
					nil,
				),
			},
			want: `gauge:<value:6.12 > `,
		},
		{
			payload: "counter-1=3.14:gauge-1=6.12:flag=false:counter-2=6.66",
			field:   "counter-1",
			metric: cmetric{
				typ: prometheus.CounterValue,
				desc: prometheus.NewDesc(
					"c1",
					"My first counter",
					nil,
					nil,
				),
			},
			want: `counter:<value:3.14 > `,
		},
	} {
		ch := make(chan prometheus.Metric)
		metrics := cmetrics{c.field: c.metric}
		go infoCollect(ch, metrics, c.payload)

		var metric prometheus.Metric
		select {
		case metric = <-ch:
		case <-time.After(time.Second):
			t.Fatalf("timeout for case %d (%q)", n, c.want)
		}

		m := &dto.Metric{}
		metric.Write(m)
		if have, want := m.String(), c.want; have != want {
			t.Errorf("case %d: have %q, want %q", n, have, want)
		}
	}
}
