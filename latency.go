package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	latencyMetrics = []string{"query", "query-rec-count", "read", "udf", "write"}
	nsHeader       = regexp.MustCompile("^{(?P<namespace>.+)}-(?P<operation>.+)$")
)

type latencyCollector struct {
	latency cmetrics
	ops     cmetrics
}

func newLatencyCollector() latencyCollector {
	lc := latencyCollector{
		latency: map[string]cmetric{},
		ops:     map[string]cmetric{},
	}
	for _, m := range latencyMetrics {
		lc.latency[m] = cmetric{
			typ: prometheus.GaugeValue,
			desc: prometheus.NewDesc(
				promkey(systemLatency, m),
				m+" latency histogram",
				[]string{"namespace", "threshold"},
				nil,
			),
		}
		lc.ops[m] = cmetric{
			typ: prometheus.GaugeValue,
			desc: prometheus.NewDesc(
				promkey(systemOps, m),
				m+" ops per second",
				[]string{"namespace"},
				nil,
			),
		}
	}
	return lc
}

func (lc latencyCollector) describe(ch chan<- *prometheus.Desc) {
	for _, s := range lc.latency {
		ch <- s.desc
	}
	for _, s := range lc.ops {
		ch <- s.desc
	}
}

func (lc latencyCollector) collect(conn *as.Connection) ([]prometheus.Metric, error) {
	stats, err := as.RequestInfo(conn, "latency:")
	if err != nil {
		return nil, err
	}
	lat, err := parseLatency(stats["latency:"])
	if err != nil {
		return nil, err
	}
	var metrics []prometheus.Metric
	for key, ms := range lat {
		if key == "batch-index" {
			continue // TODO: would be nice to do something with this key
		}
		ns, op, err := readNS(key)
		if err != nil {
			return nil, fmt.Errorf("weird latency key %q: %s", key, err)
		}
		for threshold, data := range ms {
			if threshold == "ops/sec" {
				m := lc.ops[op]
				metrics = append(
					metrics,
					prometheus.MustNewConstMetric(m.desc, m.typ, data, ns),
				)
				continue
			}
			m := lc.latency[op]
			metrics = append(
				metrics,
				prometheus.MustNewConstMetric(m.desc, m.typ, data, ns, threshold),
			)
		}
	}
	return metrics, nil
}

// parseLatency returns map with: "[{namespace}]-[op]" -> map[threshold]measurement
// It doesn't interprets the keys.
func parseLatency(lat string) (map[string]map[string]float64, error) {
	results := map[string]map[string]float64{}
	// Lines come in pairs, and look like this:
	// reads:{namespace}-read:14:08:38-GMT,ops/sec,>1ms,>8ms,>64ms;14:08:48,2586.8,1.58,0.77,0.00;
	lines := strings.Split(lat, ";")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "error") {
			continue
		}
		vs := strings.Split(line, ",")
		key := strings.SplitN(vs[0], ":", 2)[0] // strips timestamp
		cols := vs[1:]

		if i+1 >= len(lines) {
			return nil, fmt.Errorf("latency: missing measurements line")
		}
		nextLine := lines[i+1]
		i++
		measurements := strings.Split(nextLine, ",")
		if len(measurements) != len(cols)+1 {
			return nil, fmt.Errorf("invalid latency format")
		}

		ms := map[string]float64{}
		for i, v := range measurements[1:] {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("%q invalid latency value %q: %s", namespace, v, err)
			}
			ms[cols[i]] = f
		}
		results[key] = ms
	}
	return results, nil
}

// readNS converts a key like "{foo}-bar" to "foo", "bar"
func readNS(s string) (string, string, error) {
	m := nsHeader.FindStringSubmatch(s)
	if len(m) != 3 {
		return "", "", fmt.Errorf("invalid namespace key")
	}
	return m[1], m[2], nil
}
