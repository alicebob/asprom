package main

import (
	"log"
	"strings"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// namespaceMetrics lists the keys we report from aero's namespace statistics command.
	// See `asinfo -l -v namespace/<namespace>` for the full list.
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
		log.Print(err)
		return
	}
	for _, ns := range strings.Split(info["namespaces"], ";") {
		nsinfo, err := as.RequestInfo(conn, "namespace/"+ns)
		if err != nil {
			log.Print(err)
			continue
		}
		ms := map[string]setter{}
		for key, m := range c.gauges {
			ms[key] = m.WithLabelValues(ns)
		}
		infoCollect(ch, ms, nsinfo["namespace/"+ns])
	}
}
