package main

import (
	"log"
	"strings"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// StatsMetrics lists the keys we report from aero's info:statistics
	// command.
	// See `asinfo -l -v statistics` for the full list.
	StatsMetrics = []metric{
		{collGauge, "cluster_size", "cluster size, as reported by this node"},
		{collGauge, "free-pct-disk", "disk free %"},
		{collGauge, "free-pct-memory", "memory free %"},
		{collGauge, "migrate_rx_objs", "cluster wide migrate rx objects"},
		{collGauge, "migrate_tx_objs", "cluster wide migrate tx objects"},
		{collGauge, "objects", "objects per node"},
		{collGauge, "client_connections", "client connections per node"},
		{collCounter, "stat_evicted_objects", "evicted objects"},
		{collCounter, "stat_expired_objects", "expired objects"},
	}
)

type statsCollector struct {
	metrics map[string]setter
}

func newStatsCollector() *statsCollector {
	smetrics := map[string]setter{}
	for _, s := range StatsMetrics {
		key := s.aeroName
		promName := strings.Replace(key, "-", "_", -1)
		switch s.typ {
		case collGauge:
			smetrics[key] = prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: systemNode,
				Name:      promName,
				Help:      s.desc,
			})
		case collCounter:
			smetrics[key] = prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: systemNode,
				Name:      promName,
				Help:      s.desc,
			})
		}
	}

	return &statsCollector{
		metrics: smetrics,
	}
}

func (s *statsCollector) describe(ch chan<- *prometheus.Desc) {
	for _, s := range s.metrics {
		s.Describe(ch)
	}
}

func (s *statsCollector) collect(conn *as.Connection, ch chan<- prometheus.Metric) {
	res, err := as.RequestInfo(conn, "statistics")
	if err != nil {
		log.Print(err)
		return
	}
	infoCollect(ch, s.metrics, res["statistics"])
}
