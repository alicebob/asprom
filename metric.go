package main

import (
	"fmt"
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

func parseFloatOrBool(v string) (float64, error) {
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f, nil
	}

	if b, err := strconv.ParseBool(v); err == nil {
		if b {
			return 1, nil
		}
		return 0, nil
	}

	return 0, fmt.Errorf("not a float or bool: %q", v)
}

// infoCollect parses RequestInfo() results and handles the metrics
func infoCollect(
	metrics cmetrics,
	info string,
	labelValues ...string,
) []prometheus.Metric {
	var res []prometheus.Metric
	stats := parseInfo(info)
	for key, m := range metrics {
		v, ok := stats[key]
		if !ok {
			// key presence depends on (namespace) configuration
			continue
		}
		f, err := parseFloatOrBool(v)
		if err != nil {
			log.Printf("%q invalid value %q: %s", key, v, err)
			continue
		}
		res = append(
			res,
			prometheus.MustNewConstMetric(m.desc, m.typ, f, labelValues...),
		)
	}
	return res
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
	replacer := strings.NewReplacer("-", "_", ".", "_")
	k := replacer.Replace(key)
	return namespace + "_" + sys + "_" + k
}
