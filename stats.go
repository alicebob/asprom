package main

import (
	"log"
	"strings"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// StatsMetrics lists the keys we report from aero's info:statistics
	// command.
	// See `asinfo -l -v statistics` for the full list.
	StatsMetrics = []metric{
                {collGauge, "cluster_size", "cluster size"},
                {collGauge, "system_free_mem_pct", "system free mem pct"},
                {collGauge, "objects", "objects"},
                {collGauge, "sub_objects", "sub objects"},
                {collGauge, "info_queue", "info queue"},
                {collGauge, "delete_queue", "delete queue"},
                {collGauge, "client_connections", "client connections"},
                {collGauge, "heartbeat_connections", "heartbeat connections"},
                {collGauge, "fabric_connections", "fabric connections"},
                {collGauge, "batch_index_initiate", "batch index initiate"},
                {collGauge, "batch_index_complete", "batch index complete"},
                {collGauge, "batch_index_error", "batch index error"},
                {collGauge, "batch_index_timeout", "batch index timeout"},
                {collGauge, "batch_initiate", "batch initiate"},
                {collGauge, "batch_queue", "batch queue"},
                {collGauge, "batch_error", "batch error"},
                {collGauge, "batch_timeout", "batch timeout"},
                {collGauge, "scans_active", "scans active"},
                {collGauge, "query_short_running", "query short running"},
                {collGauge, "query_long_running", "query long running"},
                {collGauge, "sindex_ucgarbage_found", "sindex ucgarbage found"},
                {collGauge, "sindex_gc_locktimedout", "sindex gc locktimedout"},
                {collGauge, "sindex_gc_inactivity_dur", "sindex gc inactivity dur"},
                {collGauge, "sindex_gc_activity_dur", "sindex gc activity dur"},
                {collGauge, "sindex_gc_list_creation_time", "sindex gc list creation time"},
                {collGauge, "sindex_gc_list_deletion_time", "sindex gc list deletion time"},
                {collGauge, "sindex_gc_objects_validated", "sindex gc objects validated"},
                {collGauge, "sindex_gc_garbage_found", "sindex gc garbage found"},
                {collGauge, "sindex_gc_garbage_cleaned", "sindex gc garbage cleaned"},
                {collCounter, "fabric_msgs_sent", "fabric msgs sent"},
                {collCounter, "fabric_msgs_rcvd", "fabric msgs rcvd"},
                {collCounter, "xdr_ship_success", "xdr ship success"},
                {collCounter, "xdr_ship_delete_success", "xdr ship delete success"},
                {collCounter, "xdr_ship_source_error", "xdr ship source error"},
                {collCounter, "xdr_ship_destination_error", "xdr ship destination error"},
                {collGauge, "xdr_ship_bytes", "xdr ship bytes"},
                {collGauge, "xdr_ship_latency_avg", "xdr ship latency avg"},
                {collGauge, "xdr_ship_compression_avg_pct", "xdr ship compression avg pct"},
                {collGauge, "xdr_ship_inflight_objects", "xdr ship inflight objects"},
                {collGauge, "xdr_ship_outstanding_objects", "xdr ship outstanding objects"},
                {collCounter, "xdr_read_success", "xdr read success"},
                {collCounter, "xdr_read_error", "xdr read error"},
                {collGauge, "xdr_read_notfound", "xdr read notfound"},
                {collGauge, "xdr_read_latency_avg", "xdr read latency avg"},
                {collGauge, "xdr_read_active_avg_pct", "xdr read active avg pct"},
                {collGauge, "xdr_read_idle_avg_pct", "xdr read idle avg pct"},
                {collGauge, "xdr_read_reqq_used", "xdr read reqq used"},
                {collGauge, "xdr_read_reqq_used_pct", "xdr read reqq used pct"},
                {collGauge, "xdr_read_respq_used", "xdr read respq used"},
                {collGauge, "xdr_read_txnq_used", "xdr read txnq used"},
                {collGauge, "xdr_read_txnq_used_pct", "xdr read txnq used pct"},
                {collGauge, "xdr_queue_overflow_error", "xdr queue overflow error"},
                {collGauge, "xdr_hotkey_fetch", "xdr hotkey fetch"},
                {collGauge, "xdr_hotkey_skip", "xdr hotkey skip"},
                {collCounter, "xdr_unknown_namespace_error", "xdr unknown namespace error"},
                {collCounter, "xdr_uninitialized_destination_error", "xdr uninitialized destination error"},
                {collGauge, "xdr_timelag", "xdr timelag"},
                {collGauge, "xdr_throughput", "xdr throughput"},
	}
)

type statsCollector struct {
	metrics map[string]setter
}

func newStatsCollector() *statsCollector {
	smetrics := map[string]setter{}
	for _, s := range StatsMetrics {
		key := s.aeroName
		promName := strings.Replace(key, "-", "_", -1)
		switch s.typ {
		case collGauge:
			smetrics[key] = prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: systemNode,
				Name:      promName,
				Help:      s.desc,
			})
		case collCounter:
			smetrics[key] = prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: systemNode,
				Name:      promName,
				Help:      s.desc,
			})
		}
	}

	return &statsCollector{
		metrics: smetrics,
	}
}

func (s *statsCollector) describe(ch chan<- *prometheus.Desc) {
	for _, s := range s.metrics {
		s.Describe(ch)
	}
}

func (s *statsCollector) collect(conn *as.Connection, ch chan<- prometheus.Metric) {
	res, err := as.RequestInfo(conn, "statistics")
	if err != nil {
		log.Print(err)
		return
	}
	infoCollect(ch, s.metrics, res["statistics"])
}
