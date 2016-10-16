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
                {collGauge, "objects", "objects"},
                {collGauge, "sub_objects", "sub objects"},
                {collGauge, "non_expirable_objects", "non expirable objects"},
                {collCounter, "expired_objects", "expired objects"},
                {collCounter, "evicted_objects", "evicted objects"},
                {collCounter, "set_deleted_objects", "set deleted objects"},
                {collGauge, "memory_used_bytes", "memory used bytes"},
                {collGauge, "memory_used_data_bytes", "memory used data bytes"},
                {collGauge, "memory_used_index_bytes", "memory used index bytes"},
                {collGauge, "memory_used_sindex_bytes", "memory used sindex bytes"},
                {collGauge, "memory_free_pct", "memory free pct"},
                {collGauge, "device_total_bytes", "device total bytes"},
                {collGauge, "device_used_bytes", "device used bytes"},
                {collGauge, "device_free_pct", "device free pct"},
                {collGauge, "device_available_pct", "device available pct"},
                {collGauge, "cache_read_pct", "cache read pct"},
                {collGauge, "migrate_records_skipped", "migrate records skipped"},
                {collGauge, "migrate_records_transmitted", "migrate records transmitted"},
                {collGauge, "migrate_record_retransmits", "migrate record retransmits"},
                {collGauge, "migrate_record_receives", "migrate record receives"},
                {collCounter, "client_proxy_complete", "client proxy complete"},
                {collCounter, "client_proxy_error", "client proxy error"},
                {collCounter, "client_proxy_timeout", "client proxy timeout"},
                {collCounter, "client_read_success", "client read success"},
                {collCounter, "client_read_error", "client read error"},
                {collCounter, "client_read_timeout", "client read timeout"},
                {collCounter, "client_read_not_found", "client read not found"},
                {collCounter, "client_write_success", "client write success"},
                {collCounter, "client_write_error", "client write error"},
                {collCounter, "client_write_timeout", "client write timeout"},
                {collCounter, "xdr_write_success", "xdr write success"},
                {collCounter, "xdr_write_error", "xdr write error"},
                {collCounter, "xdr_write_timeout", "xdr write timeout"},
                {collCounter, "client_delete_success", "client delete success"},
                {collCounter, "client_delete_error", "client delete error"},
                {collCounter, "client_delete_timeout", "client delete timeout"},
                {collCounter, "client_delete_not_found", "client delete not found"},
                {collCounter, "client_lang_read_success", "client lang read success"},
                {collCounter, "client_lang_write_success", "client lang write success"},
                {collCounter, "client_lang_delete_success", "client lang delete success"},
                {collCounter, "client_lang_error", "client lang error"},
                {collCounter, "batch_sub_proxy_complete", "batch sub proxy complete"},
                {collCounter, "batch_sub_proxy_error", "batch sub proxy error"},
                {collCounter, "batch_sub_proxy_timeout", "batch sub proxy timeout"},
                {collCounter, "batch_sub_read_success", "batch sub read success"},
                {collCounter, "batch_sub_read_error", "batch sub read error"},
                {collCounter, "batch_sub_read_timeout", "batch sub read timeout"},
                {collCounter, "batch_sub_read_not_found", "batch sub read not found"},
                {collCounter, "scan_basic_complete", "scan basic complete"},
                {collCounter, "scan_basic_error", "scan basic error"},
                {collCounter, "scan_basic_abort", "scan basic abort"},
                {collCounter, "query_reqs", "query reqs"},
                {collCounter, "query_fail", "query fail"},
                {collCounter, "query_short_queue_full", "query short queue full"},
                {collCounter, "query_long_queue_full", "query long queue full"},
                {collCounter, "query_short_reqs", "query short reqs"},
                {collCounter, "query_long_reqs", "query long reqs"},
                {collCounter, "query_agg", "query agg"},
                {collCounter, "query_agg_success", "query agg success"},
                {collCounter, "query_agg_error", "query agg error"},
                {collCounter, "query_agg_abort", "query agg abort"},
                {collCounter, "query_agg_avg_rec_count", "query agg avg rec count"},
                {collCounter, "query_lookups", "query lookups"},
                {collCounter, "query_lookup_success", "query lookup success"},
                {collCounter, "query_lookup_error", "query lookup error"},
                {collCounter, "query_lookup_abort", "query lookup abort"},
                {collCounter, "query_lookup_avg_rec_count", "query lookup avg rec count"},
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
