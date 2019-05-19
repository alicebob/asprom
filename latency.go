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
	latency          cmetrics
	latencyHistogram cmetrics
	ops              cmetrics
	histOps          cmetrics
	bucketSum        cmetrics
}

func newLatencyCollector() latencyCollector {
	lc := latencyCollector{
		latency:          map[string]cmetric{},
		latencyHistogram: map[string]cmetric{},
		ops:              map[string]cmetric{},
		histOps:          map[string]cmetric{},
		bucketSum:        map[string]cmetric{},
	}
	for _, m := range latencyMetrics {
		lc.latency[m] = cmetric{
			typ: prometheus.GaugeValue,
			desc: prometheus.NewDesc(
				promkey(systemLatency, m),
				m+" latency",
				[]string{"namespace", "threshold"}, // threshold to be printed as le for histogram
				nil,
			),
		}
		lc.latencyHistogram[m] = cmetric{
			typ: prometheus.GaugeValue,
			desc: prometheus.NewDesc(
				// for prom histogram latency buckets, metric must end in _bucket
				promkey(systemLatencyHist, m+"_bucket"),
				m+" latency histogram",
				[]string{"namespace", "le"}, // threshold to be printed as le for histogram, le="1" means ops that completed in less than 1ms
				nil,
			),
		}
		lc.histOps[m] = cmetric{
			typ: prometheus.GaugeValue,
			desc: prometheus.NewDesc(
				// for prom histogram, must have a metric ending in _count which is equal to the sum of all observed events
				promkey(systemLatencyHist, m+"_count"),
				m+" ops per second for histogram",
				[]string{"namespace"},
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
		lc.bucketSum[m] = cmetric{
			typ: prometheus.GaugeValue,
			desc: prometheus.NewDesc(
				// for prom histogram, must have a metric ending in _sum which is equal to the sum of all observed events values
				promkey(systemLatencyHist, m+"_sum"),
				m+" sum of all buckets",
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
	for _, s := range lc.histOps {
		ch <- s.desc
	}
	for _, s := range lc.bucketSum {
		ch <- s.desc
	}
	for _, s := range lc.latencyHistogram {
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
	re := regexp.MustCompile("[0-9.]+") // regex to pull the number from the bucket name, >1ms -> 1, >8ms -> 8 etc.

	for key, ms := range lat {
		if key == "batch-index" {
			continue // TODO: would be nice to do something with this key
		}
		ns, op, err := readNS(key)
		if err != nil {
			return nil, fmt.Errorf("weird latency key %q: %s", key, err)
		}
		// need to grab ops outside of the latency loop
		// so that we can use it for estimatedBucketOps later
		// the latency map could be out of order, so OPS/S needs to be accessed first
		ops := ms["ops/sec"]
		var bucketSum float64
		histOpsMetric := lc.histOps[op]
		metrics = append(
			metrics,
			prometheus.MustNewConstMetric(histOpsMetric.desc, histOpsMetric.typ, ops, ns),
		)

		opsMetric := lc.ops[op]
		metrics = append(
			metrics,
			prometheus.MustNewConstMetric(opsMetric.desc, opsMetric.typ, ops, ns),
		)

		for threshold, data := range ms {
			if threshold == "ops/sec" {
				continue
			}
			thresholdNum := re.FindString(threshold) // filter out >1ms to just the number 1, similarly >8ms becomes 8..
			bucketVal, _ := strconv.ParseFloat(thresholdNum, 64)
			m := lc.latencyHistogram[op]
			// latency is exported as % in certain buckets.
			// For histogram consumption, it would be nice to have the estimated number of
			// operations in each bucket instead. So this is just some simple math to figure out, given the total ops/s and % in each bucket
			// How many ops in each bucket.
			estimatedBucketOps := ops * data / 100.0
			bucketSum += (bucketVal * estimatedBucketOps) // to generate sum like aerospike_latency_hist_read_sum
			if bucketVal == 1.0 {
				below1msAssumption := ops - estimatedBucketOps
				bucketSum += 0.5 * below1msAssumption // for the blahblah_sum histogram metric, to calculate averages, assume the transactions <1ms are .5ms
			}
			leBucketOps := ops - estimatedBucketOps
			metrics = append(
				metrics,
				prometheus.MustNewConstMetric(m.desc, m.typ, leBucketOps, ns, thresholdNum),
			)
			m = lc.latency[op]
			metrics = append(
				metrics,
				prometheus.MustNewConstMetric(m.desc, m.typ, data, ns, threshold),
			)
		}
		m := lc.bucketSum[op]
		metrics = append(
			metrics,
			prometheus.MustNewConstMetric(m.desc, m.typ, bucketSum, ns),
		)
		m = lc.latencyHistogram[op]
		metrics = append(
			metrics,
			prometheus.MustNewConstMetric(m.desc, m.typ, ops, ns, "+Inf"),
		)
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
