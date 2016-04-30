package main

import (
	"log"
	"strconv"
	"strings"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

type collType int

const (
	collGauge collType = iota
	collCounter
)

type metric struct {
	typ      collType
	aeroName string
	desc     string
}

var (
	// statsMetrics lists the keys we report from aero's info:statistics
	// command.
	// See `asinfo -v statistics` for the full list.
	statsMetrics = []metric{
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

// setter is a Gauge or a Counter
type setter interface {
	prometheus.Metric
	prometheus.Collector
	Set(float64)
}

type statsCollector struct {
	metrics map[string]setter
}

func newStatsCollector() *statsCollector {
	smetrics := map[string]setter{}
	for _, s := range statsMetrics {
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
		// TODO
		log.Print(err)
		return
	}
	stats := parseInfo(res["statistics"])

	for key, m := range s.metrics {
		v, ok := stats[key]
		if !ok {
			log.Printf("key %q not present. Typo?", key)
			continue
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Printf("%q invalid value %q: %s", key, v, err)
			continue
		}
		m.Set(f)
		ch <- m
	}
}

func parseInfo(s string) map[string]string {
	r := map[string]string{}
	for _, v := range strings.Split(s, ";") {
		kv := strings.Split(v, "=")
		if len(kv) > 1 {
			r[kv[0]] = kv[1]
		}
	}
	return r
}
