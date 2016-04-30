package main

import (
	"log"
	"strconv"
	"strings"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// namespaceMetrics lists the keys we report from aero's namespace statistics command.
	// See `asinfo -v namespace/<namespace>` for the full list.
	namespaceMetrics = []metric{
		{collGauge, "migrate-rx-partitions-remaining", "remaining rx migrate partitions per namespace per node"},
		{collGauge, "migrate-tx-partitions-remaining", "remaining tx migrate partitions per namespace per node"},
		{collGauge, "free-pct-memory", "% free memory per namespace per node"},
		{collGauge, "evicted-objects", "evicted objects per namespace per node"},
		{collGauge, "expired-objects", "expired objects per namespace per node"},
		{collGauge, "objects", "objects per namespace per node"},
	}
)

type nsCollector struct {
	gauges map[string]*prometheus.GaugeVec
}

func newNSCollector() *nsCollector {
	gauges := map[string]*prometheus.GaugeVec{}
	for _, s := range namespaceMetrics {
		key := s.aeroName
		promName := strings.Replace(key, "-", "_", -1)
		switch s.typ {
		case collGauge:
			gauges[key] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: systemNamespace,
				Name:      promName,
				Help:      s.desc,
			},
				[]string{"namespace"},
			)
		case collCounter:
			// todo
		}
	}

	return &nsCollector{
		gauges: gauges,
	}
}

func (c *nsCollector) describe(ch chan<- *prometheus.Desc) {
	for _, s := range c.gauges {
		s.Describe(ch)
	}
}

func (c *nsCollector) collect(conn *as.Connection, ch chan<- prometheus.Metric) {
	info, err := as.RequestInfo(conn, "namespaces")
	if err != nil {
		// TODO
		log.Print(err)
		return
	}
	// log.Printf("namespaces: %+v\n", info["namespaces"])
	for _, ns := range strings.Split(info["namespaces"], ";") {
		nsinfo, err := as.RequestInfo(conn, "namespace/"+ns)
		if err != nil {
			// TODO
			log.Print(err)
			return
		}
		stats := parseInfo(nsinfo["namespace/"+ns])
		for key, m := range c.gauges {
			v, ok := stats[key]
			if !ok {
				log.Printf("ns key %q not present. Typo?", key)
				continue
			}
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				log.Printf("%q invalid value %q: %s", key, v, err)
				continue
			}
			mi := m.WithLabelValues(ns)
			mi.Set(f)
			ch <- mi
		}
	}
}
