package main

import (
	"log"
	"strconv"
	"strings"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	latencyMetrics = []string{"reads", "writes_master", "proxy", "udf", "query"}
	// should match the columns from `asinfo -v "latency:"`
	latencyIntervals = []string{">1ms", ">8ms", ">64ms"}
)

type latencyCollector struct {
	metrics map[string]prometheus.Gauge
}

func newLatencyCollector() *latencyCollector {
	s := &latencyCollector{
		metrics: map[string]prometheus.Gauge{},
	}
	for _, m := range latencyMetrics {
		s.metrics[m+"_ops_sec"] = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: systemLatency,
			Name:      m + "_ops_sec",
			Help:      m + " ops per second",
		})
		for _, int := range latencyIntervals {
			promName := strings.Replace(m+"_"+int, ">", "gt_", -1)
			s.metrics[m+"_"+int] = prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: systemLatency,
				Name:      promName,
				Help:      m + " " + int,
			})
		}
	}
	return s
}

func (s *latencyCollector) describe(ch chan<- *prometheus.Desc) {
	for _, m := range s.metrics {
		m.Describe(ch)
	}
}

func (s *latencyCollector) collect(conn *as.Connection, ch chan<- prometheus.Metric) {
	stats, err := as.RequestInfo(conn, "latency:")
	if err != nil {
		// TODO
		log.Print(err)
		return
	}
	lines := strings.Split(stats["latency:"], ";")
	// Lines come in pairs, and look like this:
	// reads:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,12469.3,0.40,0.00,0.00;writes_master:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,0.0,0.00,0.00,0.00;proxy:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,0.0,0.00,0.00,0.00;udf:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,14730.7,0.03,0.00,0.00;query:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,0.0,0.00,0.00,0.00;
	for len(lines) >= 2 {
		first := strings.SplitN(lines[0], ":", 2)
		if len(first) != 2 {
			log.Print("invalid latency format")
			return
		}
		typ := first[0]
		headers := strings.Split(first[1], ",")
		values := strings.Split(lines[1], ",")
		lines = lines[2:]
		for i, h := range headers[1:] {
			v := values[i+1]
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				log.Printf("%q invalid latency value %q: %s", h, v, err)
				continue
			}
			switch {
			case h == "ops/sec":
				if m, ok := s.metrics[typ+"_ops_sec"]; ok {
					m.Set(f)
					ch <- m
				} else {
					log.Printf("unknown latency type: %q", typ)
				}
			case strings.HasPrefix(h, ">"):
				if m, ok := s.metrics[typ+"_"+h]; ok {
					m.Set(f)
					ch <- m
				}
			}
		}
	}
}
