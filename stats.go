package main

import (
	"log"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// StatsMetrics lists the keys we report from aero's info:statistics
	// command.
	// See `asinfo -l -v statistics` for the full list.
	StatsMetrics = []metric{
		gauge("cluster_size", "cluster size"),
		gauge("system_free_mem_pct", "system free mem pct"),
		gauge("objects", "objects"),
		gauge("sub_objects", "sub objects"),
		gauge("info_queue", "info queue"),
		gauge("delete_queue", "delete queue"),
		gauge("client_connections", "client connections"),
		gauge("heartbeat_connections", "heartbeat connections"),
		gauge("fabric_connections", "fabric connections"),
		gauge("batch_index_initiate", "batch index initiate"),
		gauge("batch_index_complete", "batch index complete"),
		gauge("batch_index_error", "batch index error"),
		gauge("batch_index_timeout", "batch index timeout"),
		gauge("batch_initiate", "batch initiate"),
		gauge("batch_queue", "batch queue"),
		gauge("batch_error", "batch error"),
		gauge("batch_timeout", "batch timeout"),
		gauge("scans_active", "scans active"),
		gauge("query_short_running", "query short running"),
		gauge("query_long_running", "query long running"),
		gauge("sindex_ucgarbage_found", "sindex ucgarbage found"),
		gauge("sindex_gc_locktimedout", "sindex gc locktimedout"),
		gauge("sindex_gc_inactivity_dur", "sindex gc inactivity dur"),
		gauge("sindex_gc_activity_dur", "sindex gc activity dur"),
		gauge("sindex_gc_list_creation_time", "sindex gc list creation time"),
		gauge("sindex_gc_list_deletion_time", "sindex gc list deletion time"),
		gauge("sindex_gc_objects_validated", "sindex gc objects validated"),
		gauge("sindex_gc_garbage_found", "sindex gc garbage found"),
		gauge("sindex_gc_garbage_cleaned", "sindex gc garbage cleaned"),
		counter("fabric_msgs_sent", "fabric msgs sent"),
		counter("fabric_msgs_rcvd", "fabric msgs rcvd"),
		counter("xdr_ship_success", "xdr ship success"),
		counter("xdr_ship_delete_success", "xdr ship delete success"),
		counter("xdr_ship_source_error", "xdr ship source error"),
		counter("xdr_ship_destination_error", "xdr ship destination error"),
		gauge("xdr_ship_bytes", "xdr ship bytes"),
		gauge("xdr_ship_latency_avg", "xdr ship latency avg"),
		gauge("xdr_ship_compression_avg_pct", "xdr ship compression avg pct"),
		gauge("xdr_ship_inflight_objects", "xdr ship inflight objects"),
		gauge("xdr_ship_outstanding_objects", "xdr ship outstanding objects"),
                counter("xdr_write_success", "xdr write success"),
                counter("xdr_write_error", "xdr write error"),
                counter("xdr_write_timeout", "xdr write timeout"),
		counter("xdr_read_success", "xdr read success"),
		counter("xdr_read_error", "xdr read error"),
		gauge("xdr_read_notfound", "xdr read notfound"),
		gauge("xdr_read_latency_avg", "xdr read latency avg"),
		gauge("xdr_read_active_avg_pct", "xdr read active avg pct"),
		gauge("xdr_read_idle_avg_pct", "xdr read idle avg pct"),
		gauge("xdr_read_reqq_used", "xdr read reqq used"),
		gauge("xdr_read_reqq_used_pct", "xdr read reqq used pct"),
		gauge("xdr_read_respq_used", "xdr read respq used"),
		gauge("xdr_read_txnq_used", "xdr read txnq used"),
		gauge("xdr_read_txnq_used_pct", "xdr read txnq used pct"),
		gauge("xdr_queue_overflow_error", "xdr queue overflow error"),
		gauge("xdr_hotkey_fetch", "xdr hotkey fetch"),
		gauge("xdr_hotkey_skip", "xdr hotkey skip"),
		counter("xdr_unknown_namespace_error", "xdr unknown namespace error"),
		counter("xdr_uninitialized_destination_error", "xdr uninitialized destination error"),
		gauge("xdr_timelag", "xdr timelag"),
		gauge("xdr_throughput", "xdr throughput"),
	}
)

type statsCollector cmetrics

func newStatsCollector() statsCollector {
	smetrics := map[string]cmetric{}
	for _, m := range StatsMetrics {
		smetrics[m.aeroName] = cmetric{
			desc: prometheus.NewDesc(
				promkey(systemNode, m.aeroName),
				m.desc,
				nil,
				nil,
			),
			typ: m.typ,
		}
	}
	return smetrics
}

func (sc statsCollector) describe(ch chan<- *prometheus.Desc) {
	for _, s := range sc {
		ch <- s.desc
	}
}

func (sc statsCollector) collect(conn *as.Connection, ch chan<- prometheus.Metric) {
	res, err := as.RequestInfo(conn, "statistics")
	if err != nil {
		log.Print(err)
		return
	}
	infoCollect(ch, cmetrics(sc), res["statistics"])
}
