package main

import (
	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// StatsMetrics lists the keys we report from aero's info:statistics
	// command.
	// See `asinfo -l -v statistics` for the full list.
	StatsMetrics = []metric{
		gauge("cluster_size", "cluster size"),
		// cluster_key=C0758EC6A81F
		// cluster_integrity=true
		// cluster_is_member=true
		counter("uptime", "uptime"),
		gauge("system_free_mem_pct", "system free mem pct"),
		// system_swapping=false
		gauge("heap_allocated_kbytes", "heap allocated kbytes"),
		gauge("heap_active_kbytes", "heap active kbytes"),
		gauge("heap_mapped_kbytes", "heap mapped kbytes"),
		gauge("heap_efficiency_pct", "heap efficiency pct"),
		gauge("heap_site_count", "heap site count"),
		gauge("objects", "objects"),
		gauge("tombstones", "tombstones"),
		gauge("tsvc_queue", "tsvc queue"),
		gauge("info_queue", "info queue"),
		gauge("delete_queue", "delete queue"),
		// rw_in_progress=0
		// proxy_in_progress=0
		// tree_gc_queue=0
		gauge("client_connections", "client connections"),
		gauge("heartbeat_connections", "heartbeat connections"),
		gauge("fabric_connections", "fabric connections"),
		counter("heartbeat_received_self", "heartbeat received self"),
		counter("heartbeat_received_foreign", "heartbeat received foreign"),
		counter("reaped_fds", "reaped fds"),
		counter("info_complete", "info complete"),
		counter("demarshal_error", "demarshal error"),
		counter("early_tsvc_client_error", "early tsvc client error"),
		counter("early_tsvc_batch_sub_error", "early tsvc batch sub error"),
		counter("early_tsvc_udf_sub_error", "early tsvc udf sub error"),
		gauge("batch_index_initiate", "batch index initiate"),
		// batch_index_queue=0:0,0:0,0:0,0:0
		gauge("batch_index_complete", "batch index complete"),
		gauge("batch_index_error", "batch index error"),
		gauge("batch_index_timeout", "batch index timeout"),
		gauge("batch_index_unused_buffers", "batch index unused buffers"),
		gauge("batch_index_huge_buffers", "batch index huge buffers"),
		counter("batch_index_created_buffers", "batch index created buffers"),
		counter("batch_index_destroyed_buffers", "batch index destroyed buffers"),
		gauge("batch_initiate", "batch initiate"),
		gauge("batch_queue", "batch queue"),
		gauge("batch_error", "batch error"),
		gauge("batch_timeout", "batch timeout"),
		gauge("scans_active", "scans active"),
		gauge("query_short_running", "query short running"),
		gauge("query_long_running", "query long running"),
		gauge("sindex_ucgarbage_found", "sindex ucgarbage found"),
		gauge("sindex_gc_locktimedout", "sindex gc locktimedout"),
		gauge("sindex_gc_list_creation_time", "sindex gc list creation time"),
		gauge("sindex_gc_list_deletion_time", "sindex gc list deletion time"),
		gauge("sindex_gc_objects_validated", "sindex gc objects validated"),
		gauge("sindex_gc_garbage_found", "sindex gc garbage found"),
		gauge("sindex_gc_garbage_cleaned", "sindex gc garbage cleaned"),
		// paxos_principal=BB9508FED001500
		// migrate_allowed=true
		gauge("migrate_partitions_remaining", "migrate partitions remaining"),
		gauge("fabric_bulk_send_rate", "fabric bulk send rate"),
		gauge("fabric_bulk_recv_rate", "fabric bulk recv rate"),
		gauge("fabric_ctrl_send_rate", "fabric ctrl send rate"),
		gauge("fabric_ctrl_recv_rate", "fabric ctrl recv rate"),
		gauge("fabric_meta_send_rate", "fabric meta send rate"),
		gauge("fabric_meta_recv_rate", "fabric meta recv rate"),
		gauge("fabric_rw_send_rate", "fabric rw send rate"),
		gauge("fabric_rw_recv_rate", "fabric rw recv rate"),
		// XDR specific metrics
		// requires Aerospike EE
		gauge("dlog_free_pct", "dlog free pct"),
		counter("dlog_logged", "dlog logged"),
		counter("dlog_overwritten_error", "dlog overwritten error"),
		counter("dlog_processed_link_down", "dlog processed link down"),
		counter("dlog_processed_main", "dlog processed main"),
		counter("dlog_processed_replica", "dlog processed replica"),
		counter("dlog_relogged", "dlog relogged"),
		gauge("dlog_used_objects", "dlog used objects"),
		counter("local_recs_migration_retry", "Number of records missing in a batch call"),
		counter("stat_pipe_reads_diginfo", "Number of digest information read from the named pipe."),
		gauge("xdr_active_failed_node_sessions", "Number of active failed node sessions pending."),
		gauge("xdr_active_link_down_sessions", "Number of active link down sessions pending."),
		gauge("xdr_global_lastshiptime", "The minimum last ship time in millisecond (epoch) for XDR for across the cluster."),
		counter("xdr_hotkey_fetch", "xdr hotkey fetch"),
		counter("xdr_hotkey_skip", "xdr hotkey skip"),
		counter("xdr_queue_overflow_error", "xdr queue overflow error"),
		gauge("xdr_read_active_avg_pct", "xdr read active avg pct"),
		counter("xdr_read_error", "xdr read error"),
		gauge("xdr_read_idle_avg_pct", "xdr read idle avg pct"),
		gauge("xdr_read_latency_avg", "xdr read latency avg"),
		counter("xdr_read_notfound", "xdr read notfound"),
		gauge("xdr_read_reqq_used", "xdr read reqq used"),
		gauge("xdr_read_reqq_used_pct", "xdr read reqq used pct"),
		gauge("xdr_read_respq_used", "xdr read respq used"),
		counter("xdr_read_success", "xdr read success"),
		gauge("xdr_read_txnq_used", "xdr read txnq used"),
		gauge("xdr_read_txnq_used_pct", "xdr read txnq used pct"),
		counter("xdr_relogged_incoming", "Number of records relogged into this node's digest log by another node."),
		counter("xdr_relogged_outgoing", "Number of records relogged to another node's digest log. "),
		counter("xdr_ship_bytes", "xdr ship bytes"),
		gauge("xdr_ship_compression_avg_pct", "xdr ship compression avg pct"),
		counter("xdr_ship_delete_success", "xdr ship delete success"),
		counter("xdr_ship_destination_error", "xdr ship destination error"),
		counter("xdr_ship_destination_permanent_error", "xdr ship destination permanent error"),
		gauge("xdr_ship_fullrecord", "Number of records that did not take advantage of bin level shipping."),
		gauge("xdr_ship_inflight_objects", "xdr ship inflight objects"),
		gauge("xdr_ship_latency_avg", "xdr ship latency avg"),
		gauge("xdr_ship_outstanding_objects", "xdr ship outstanding objects"),
		counter("xdr_ship_source_error", "xdr ship source error"),
		counter("xdr_ship_success", "xdr ship success"),
		gauge("xdr_throughput", "xdr throughput"),
		gauge("xdr_timelag", "xdr timelag"),
		counter("xdr_uninitialized_destination_error", "xdr uninitialized destination error"),
		counter("xdr_unknown_namespace_error", "xdr unknown namespace error"),
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

func (sc statsCollector) collect(conn *as.Connection) ([]prometheus.Metric, error) {
	res, err := as.RequestInfo(conn, "statistics")
	if err != nil {
		return nil, err
	}
	return infoCollect(cmetrics(sc), res["statistics"]), nil
}
