package main

import (
	"strings"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// SetMetrics lists the keys we report from aero's sets
	// command.
	// See `asinfo -l -v sets` for the full list.
	SetMetrics = []metric{
		gauge("memory_data_bytes", "memory data bytes"),
		gauge("objects", "objects"),
		gauge("truncate_lut", "The most covering truncate_lut for this set"),
		counter("stop-writes-count", "stop writes count"),
	}
)

type setCollector cmetrics

func newSetCollector() setCollector {
	set := map[string]cmetric{}
	for _, m := range SetMetrics {
		set[m.aeroName] = cmetric{
			typ: m.typ,
			desc: prometheus.NewDesc(
				promkey(systemSet, m.aeroName),
				m.desc,
				[]string{"namespace", "set"},
				nil,
			),
		}
	}
	return set
}

func (setc setCollector) describe(ch chan<- *prometheus.Desc) {
	for _, s := range setc {
		ch <- s.desc
	}
}

func (setc setCollector) collect(conn *as.Connection) ([]prometheus.Metric, error) {
	var metrics []prometheus.Metric
	info, err := as.RequestInfo(conn, "sets")
	if err != nil {
		return nil, err
	}
	for _, setInfo := range strings.Split(info["sets"], ";") {
		if setInfo == "" {
			continue
		}
		setStats := parseInfo(setInfo)
		metrics = append(
			metrics,
			infoCollect(cmetrics(setc), setInfo, setStats["ns"], setStats["set"])...,
		)
	}
	return metrics, nil
}
