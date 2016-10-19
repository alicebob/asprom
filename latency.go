package main

import (
	"log"
	"strconv"
	"strings"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	LatencyMetrics = []string{"reads", "writes_master", "proxy", "udf", "query"}
	// should match the columns from `asinfo -v "latency:"`
	LatencyIntervals = []string{">1ms", ">8ms", ">64ms"}
)

type latencyCollector cmetrics

func newLatencyCollector() latencyCollector {
	lc := map[string]cmetric{}
	for _, m := range LatencyMetrics {
		lc[m+"_ops_sec"] = cmetric{
			typ: prometheus.GaugeValue,
			desc: prometheus.NewDesc(
				promkey(systemLatency, m+"_ops_sec"),
				m+" ops per second",
				nil,
				nil,
			),
		}
		for _, int := range LatencyIntervals {
			promName := strings.Replace(m+"_"+int, ">", "gt_", -1)
			lc[m+"_"+int] = cmetric{
				typ: prometheus.GaugeValue,
				desc: prometheus.NewDesc(
					promkey(systemLatency, promName),
					m+" "+int,
					nil,
					nil,
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
	for typ, ls := range lat {
		for k, v := range ls {
			switch {
			case k == "ops/sec":
				if m, ok := lc[typ+"_ops_sec"]; ok {
					ch <- prometheus.MustNewConstMetric(m.desc, m.typ, v)
				} else {
					log.Printf("unknown latency type: %q", typ)
				}
			case strings.HasPrefix(k, ">"):
				if m, ok := lc[typ+"_"+k]; ok {
					ch <- prometheus.MustNewConstMetric(m.desc, m.typ, v)
				}
			}
		}
	}
}

func parseLatency(lat string) map[string]map[string]float64 {
	resAll := map[string]map[string]float64{}
	// Lines come in pairs, and look like this:
	// reads:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,12469.3,0.40,0.00,0.00;writes_master:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,0.0,0.00,0.00,0.00;proxy:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,0.0,0.00,0.00,0.00;udf:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,14730.7,0.03,0.00,0.00;query:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,0.0,0.00,0.00,0.00;
	lines := strings.Split(lat, ";")
	for len(lines) >= 2 {
		first := strings.SplitN(lines[0], ":", 2)
		if len(first) != 2 {
			log.Print("invalid latency format")
			return nil
		}
		typ := first[0]
		headers := strings.Split(first[1], ",")
		values := strings.Split(lines[1], ",")
		lines = lines[2:]

		res := map[string]float64{}
		for i, h := range headers[1:] {
			v := values[i+1]
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				log.Printf("%q invalid latency value %q: %s", h, v, err)
				continue
			}
			res[h] = f
		}
		resAll[typ] = res
	}
	return resAll
}
