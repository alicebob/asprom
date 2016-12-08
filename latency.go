package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	latencyMetrics = []string{"read", "write", "udf", "query"}
	// should match operation types from `asinfo -v "latency:"`
	latencyIntervals    = []string{">1ms", ">8ms", ">64ms"}
	latencyOutputHeader = regexp.MustCompile("^{(?P<namespace>.+)}-(?P<operation>.+?):.+?,ops.sec,>1ms,>8ms,>64ms$")
)

type latencyCollector cmetrics

func newLatencyCollector() latencyCollector {
	lc := map[string]cmetric{}
	for _, m := range latencyMetrics {
		for _, int := range latencyIntervals {
			lc[m+int] = cmetric{
				typ: prometheus.GaugeValue,
				desc: prometheus.NewDesc(
					promkey(systemLatency, m),
					m+" latency histogram",
					[]string{"namespace"},
					prometheus.Labels{"gt": int[1:]},
				),
			}
		}
	}
	return lc
}

func (lc latencyCollector) describe(ch chan<- *prometheus.Desc) {
	for _, s := range lc {
		ch <- s.desc
	}
}

func (lc latencyCollector) collect(conn *as.Connection, ch chan<- prometheus.Metric) {
	stats, err := as.RequestInfo(conn, "latency:")
	if err != nil {
		log.Print(err)
		return
	}
	lat := parseLatency(stats["latency:"])
	for namespace, metrics := range lat {
		for metric, data := range metrics {
			if m, ok := lc[metric]; ok {
				ch <- prometheus.MustNewConstMetric(m.desc, m.typ, data, namespace)
			}
		}
	}
}

func parseLatency(lat string) map[string]map[string]float64 {
	var namespace, operation string
	results := map[string]map[string]float64{}
	// Lines come in pairs, and look like this:
	// reads:{namespace}-read:14:08:38-GMT,ops/sec,>1ms,>8ms,>64ms;14:08:48,2586.8,1.58,0.77,0.00;
	lines := strings.Split(lat, ";")
	for _, line := range lines {
		if !strings.HasPrefix(line, "error") {
			match := latencyOutputHeader.FindStringSubmatch(line)
			if len(match) > 0 {
				namespace = match[1]
				operation = match[2]
			} else {
				values := strings.Split(line, ",")
				if len(values) != 5 {
					log.Print("invalid latency format")
					return nil
				}
				result, ok := results[namespace]
				if !ok {
					result = map[string]float64{}
				}
				for i, v := range values[2:] {
					f, err := strconv.ParseFloat(v, 64)
					if err != nil {
						log.Printf("%q invalid latency value %q: %s", namespace, v, err)
						continue
					}
					result[operation+latencyIntervals[i]] = f
				}
				for _, item := range latencyMetrics {
					if operation == item {
						results[namespace] = result
					}
				}
			}
		}
	}
	return results
}
