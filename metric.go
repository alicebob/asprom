package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type metric struct {
	typ      prometheus.ValueType
	aeroName string
	desc     string
}

// cmetrics is promkey -> prom metric
type cmetrics map[string]cmetric
type cmetric struct {
	desc *prometheus.Desc
	typ  prometheus.ValueType
}

// infoCollect parses RequestInfo() results and handles the metrics
func infoCollect(
	ch chan<- prometheus.Metric,
	metrics cmetrics,
	info string,
	labelValues ...string,
) {
	stats := parseInfo(info)
	for key, m := range metrics {
		v, ok := stats[key]
		if !ok {
			log.Printf("key %q not present", key)
			continue
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Printf("%q invalid value %q: %s", key, v, err)
			continue
		}
		ch <- prometheus.MustNewConstMetric(m.desc, m.typ, f, labelValues...)
	}
}

func parseInfo(s string) map[string]string {
	r := map[string]string{}
	for _, l := range strings.Split(s, ";") {
		for _, v := range strings.Split(l, ":") {
			kv := strings.SplitN(v, "=", 2)
			if len(kv) > 1 {
				r[kv[0]] = kv[1]
			}
		}
	}
	return r
}

// gauge is a helper to add an aerospike metric
func gauge(name string, desc string) metric {
	return metric{
		typ:      prometheus.GaugeValue,
		aeroName: name,
		desc:     desc,
	}
}

// counter is a helper to add an aerospike metric
func counter(name string, desc string) metric {
	return metric{
		typ:      prometheus.CounterValue,
		aeroName: name,
		desc:     desc,
	}
}

// promkey makes the prom metric name out of an aerospike stat name
func promkey(sys, key string) string {
	k := strings.Replace(key, "-", "_", -1)
	return namespace + "_" + sys + "_" + k
}
