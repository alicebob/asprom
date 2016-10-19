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
		gauge("objects", "objects"),
		gauge("sub_objects", "sub objects"),
		gauge("non_expirable_objects", "non expirable objects"),
		counter("expired_objects", "expired objects"),
		counter("evicted_objects", "evicted objects"),
		counter("set_deleted_objects", "set deleted objects"),
		gauge("memory_used_bytes", "memory used bytes"),
		gauge("memory_used_data_bytes", "memory used data bytes"),
		gauge("memory_used_index_bytes", "memory used index bytes"),
		gauge("memory_used_sindex_bytes", "memory used sindex bytes"),
		gauge("memory_free_pct", "memory free pct"),
		gauge("memory-size", "memory size"),
		gauge("high-water-memory-pct", "high water memory pct"),
		gauge("stop-writes-pct", "stop writes pct"),
		gauge("high-water-disk-pct", "high water disk pct"),
		gauge("device_total_bytes", "device total bytes"),
		gauge("device_used_bytes", "device used bytes"),
		gauge("device_free_pct", "device free pct"),
		gauge("device_available_pct", "device available pct"),
		gauge("cache_read_pct", "cache read pct"),
		gauge("migrate_records_skipped", "migrate records skipped"),
		gauge("migrate_records_transmitted", "migrate records transmitted"),
		gauge("migrate_record_retransmits", "migrate record retransmits"),
		gauge("migrate_record_receives", "migrate record receives"),
		counter("client_proxy_complete", "client proxy complete"),
		counter("client_proxy_error", "client proxy error"),
		counter("client_proxy_timeout", "client proxy timeout"),
		counter("client_read_success", "client read success"),
		counter("client_read_error", "client read error"),
		counter("client_read_timeout", "client read timeout"),
		counter("client_read_not_found", "client read not found"),
		counter("client_write_success", "client write success"),
		counter("client_write_error", "client write error"),
		counter("client_write_timeout", "client write timeout"),
		counter("xdr_write_success", "xdr write success"),
		counter("xdr_write_error", "xdr write error"),
		counter("xdr_write_timeout", "xdr write timeout"),
		counter("client_delete_success", "client delete success"),
		counter("client_delete_error", "client delete error"),
		counter("client_delete_timeout", "client delete timeout"),
		counter("client_delete_not_found", "client delete not found"),
		counter("client_lang_read_success", "client lang read success"),
		counter("client_lang_write_success", "client lang write success"),
		counter("client_lang_delete_success", "client lang delete success"),
		counter("client_lang_error", "client lang error"),
		counter("batch_sub_proxy_complete", "batch sub proxy complete"),
		counter("batch_sub_proxy_error", "batch sub proxy error"),
		counter("batch_sub_proxy_timeout", "batch sub proxy timeout"),
		counter("batch_sub_read_success", "batch sub read success"),
		counter("batch_sub_read_error", "batch sub read error"),
		counter("batch_sub_read_timeout", "batch sub read timeout"),
		counter("batch_sub_read_not_found", "batch sub read not found"),
		counter("scan_basic_complete", "scan basic complete"),
		counter("scan_basic_error", "scan basic error"),
		counter("scan_basic_abort", "scan basic abort"),
		counter("query_reqs", "query reqs"),
		counter("query_fail", "query fail"),
		counter("query_short_queue_full", "query short queue full"),
		counter("query_long_queue_full", "query long queue full"),
		counter("query_short_reqs", "query short reqs"),
		counter("query_long_reqs", "query long reqs"),
		counter("query_agg", "query agg"),
		counter("query_agg_success", "query agg success"),
		counter("query_agg_error", "query agg error"),
		counter("query_agg_abort", "query agg abort"),
		counter("query_agg_avg_rec_count", "query agg avg rec count"),
		counter("query_lookups", "query lookups"),
		counter("query_lookup_success", "query lookup success"),
		counter("query_lookup_error", "query lookup error"),
		counter("query_lookup_abort", "query lookup abort"),
		counter("query_lookup_avg_rec_count", "query lookup avg rec count"),
	}
)

type nsCollector cmetrics

func newNSCollector() nsCollector {
	ns := map[string]cmetric{}
	for _, m := range NamespaceMetrics {
		ns[m.aeroName] = cmetric{
			typ: m.typ,
			desc: prometheus.NewDesc(
				promkey(systemNamespace, m.aeroName),
				m.desc,
				[]string{"namespace"},
				nil,
			),
		}
	}
	return ns
}

func (nc nsCollector) describe(ch chan<- *prometheus.Desc) {
	for _, s := range nc {
		ch <- s.desc
	}
}

func (nc nsCollector) collect(conn *as.Connection, ch chan<- prometheus.Metric) {
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
		infoCollect(ch, cmetrics(nc), nsinfo["namespace/"+ns], ns)
	}
}
