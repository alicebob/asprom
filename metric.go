package main

import (
	"log"
	"strconv"
	"strings"

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

// setter is a Gauge or a Counter
type setter interface {
	prometheus.Metric
	prometheus.Collector
	Set(float64)
}

// infoCollect parses RequestInfo() results.
// metrics is a map from aerospike stat key -> prometheus metric.
func infoCollect(ch chan<- prometheus.Metric, metrics map[string]setter, info string) {
	stats := parseInfo(info)
	for key, m := range metrics {
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
		kv := strings.SplitN(v, "=", 2)
		if len(kv) > 1 {
			r[kv[0]] = kv[1]
		}
	}
	return r
}
