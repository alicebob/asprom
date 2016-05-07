package main

import (
	"log"
	"strings"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// NamespaceMetrics lists the keys we report from aero's namespace statistics command.
	// See `asinfo -l -v namespace/<namespace>` for the full list.
	NamespaceMetrics = []metric{
		{collGauge, "migrate-rx-partitions-remaining", "remaining rx migrate partitions per namespace per node"},
		{collGauge, "migrate-tx-partitions-remaining", "remaining tx migrate partitions per namespace per node"},
		{collGauge, "free-pct-memory", "% free memory per namespace per node"},
		{collGauge, "evicted-objects", "evicted objects per namespace per node"},
		{collGauge, "expired-objects", "expired objects per namespace per node"},
		{collGauge, "objects", "objects per namespace per node"},
	}
)

type nsCollector struct {
	// gauges map[string]*prometheus.GaugeVec
	descs   []prometheus.Collector
	metrics map[string]func(ns string) setter
}

func newNSCollector() *nsCollector {
	var (
		descs   []prometheus.Collector
		metrics = map[string]func(ns string) setter{}
	)
	for _, s := range NamespaceMetrics {
		key := s.aeroName
		promName := strings.Replace(key, "-", "_", -1)
		switch s.typ {
		case collGauge:
			v := prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: systemNamespace,
				Name:      promName,
				Help:      s.desc,
			},
				[]string{"namespace"},
			)
			metrics[key] = func(ns string) setter {
				return v.WithLabelValues(ns)
			}
			descs = append(descs, v)
		case collCounter:
			// todo
		}
	}

	return &nsCollector{
		descs:   descs,
		metrics: metrics,
	}
}

func (c *nsCollector) describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descs {
		d.Describe(ch)
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
		for key, m := range c.metrics {
			ms[key] = m(ns)
		}
		infoCollect(ch, ms, nsinfo["namespace/"+ns])
	}
}
