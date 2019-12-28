package main

import (
	"bytes"
	"strings"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// NamespaceMetrics lists the keys we report from aero's namespace statistics command.
	// See `asinfo -l -v namespace/<namespace>` for the full list.
	NamespaceMetrics = []metric{
		// allow-nonxdr-writes=true
		// allow-xdr-writes=true
		// cache_read_pct=0
		// cold-start-evict-ttl=4294967295
		// conflict-resolution-policy=generation
		// current_time=257114127
		// data-in-index=false
		// default-ttl=2592000
		// deleted_last_bin=0
		// disable-write-dup-res=false
		// disallow-null-setname=false
		// enable-benchmarks-batch-sub=false
		// enable-benchmarks-read=false
		// enable-benchmarks-udf-sub=false
		// enable-benchmarks-udf=false
		// enable-benchmarks-write=false
		// enable-hist-proxy=false
		// enable-xdr=false
		// evict_ttl=0
		// geo2dsphere-within.earth-radius-meters=6371000
		// geo2dsphere-within.level-mod=1
		// geo2dsphere-within.max-cells=12
		// geo2dsphere-within.max-level=30
		// geo2dsphere-within.min-level=1
		// geo2dsphere-within.strict=true
		// max-ttl=315360000
		// migrate-order=5
		// migrate-retransmit-ms=5000
		// migrate-sleep=1
		// ns-forward-xdr-writes=false
		// nsup_cycle_duration=0
		// nsup_cycle_sleep_pct=0
		// obj-size-hist-max=100
		// partition-tree-locks=8
		// partition-tree-sprigs=64
		// rack-id=0
		// read-consistency-level-override=off
		// sets-enable-xdr=true
		// sindex.num-partitions=32
		// single-bin=false
		// storage-engine=memory
		// tomb-raider-eligible-age=86400
		// tomb-raider-period=86400
		// truncate_lut=0
		// truncated_records=0
		// write-commit-level-override=off
		// xmem_id=0
		counter("batch_sub_proxy_complete", "batch sub proxy complete"),
		counter("batch_sub_proxy_error", "batch sub proxy error"),
		counter("batch_sub_proxy_timeout", "batch sub proxy timeout"),
		counter("batch_sub_read_error", "batch sub read error"),
		counter("batch_sub_read_not_found", "batch sub read not found"),
		counter("batch_sub_read_success", "batch sub read success"),
		counter("batch_sub_read_timeout", "batch sub read timeout"),
		counter("batch_sub_tsvc_error", "batch sub tsvc error"),
		counter("batch_sub_tsvc_timeout", "batch sub tsvc timeout"),
		counter("client_delete_error", "client delete error"),
		counter("client_delete_not_found", "client delete not found"),
		counter("client_delete_success", "client delete success"),
		counter("client_delete_timeout", "client delete timeout"),
		counter("client_lang_delete_success", "client lang delete success"),
		counter("client_lang_error", "client lang error"),
		counter("client_lang_read_success", "client lang read success"),
		counter("client_lang_write_success", "client lang write success"),
		counter("client_proxy_complete", "client proxy complete"),
		counter("client_proxy_error", "client proxy error"),
		counter("client_proxy_timeout", "client proxy timeout"),
		counter("client_read_error", "client read error"),
		counter("client_read_not_found", "client read not found"),
		counter("client_read_success", "client read success"),
		counter("client_read_timeout", "client read timeout"),
		counter("client_tsvc_error", "client tsvc error"),
		counter("client_tsvc_timeout", "client tsvc timeout"),
		counter("client_udf_complete", "client udf complete"),
		counter("client_udf_error", "client udf error"),
		counter("client_udf_timeout", "client udf timeout"),
		counter("client_write_error", "client write error"),
		counter("client_write_success", "client write success"),
		counter("client_write_timeout", "client write timeout"),
		counter("evicted_objects", "evicted objects"),
		counter("expired_objects", "expired objects"),
		counter("fail_generation", "fail generation"),
		counter("fail_key_busy", "fail key busy"),
		counter("fail_record_too_big", "fail record too big"),
		counter("fail_xdr_forbidden", "fail xdr forbidden"),
		counter("from_proxy_delete_error", "from proxy delete error"),
		counter("from_proxy_delete_not_found", "from proxy delete not found"),
		counter("from_proxy_delete_success", "from_proxy delete success"),
		counter("from_proxy_delete_timeout", "from proxy delete timeout"),
		counter("from_proxy_read_error", "from proxy read error"),
		counter("from_proxy_read_not_found", "from proxy read not found"),
		counter("from_proxy_read_success", "from proxy read success"),
		counter("from_proxy_read_timeout", "from proxy read timeout"),
		counter("from_proxy_tsvc_error", "from proxy tsvc error"),
		counter("from_proxy_tsvc_timeout", "from proxy tsvc timeout"),
		counter("from_proxy_write_error", "from proxy write error"),
		counter("from_proxy_write_success", "from proxy write success"),
		counter("from_proxy_write_timeout", "from proxy write timeout"),
		counter("geo_region_query_cells", "geo region query cells"),
		counter("geo_region_query_falsepos", "geo region query falsepos"),
		counter("geo_region_query_points", "geo region query points"),
		counter("geo_region_query_reqs", "geo region query reqs"),
		counter("query_agg_abort", "query agg abort"),
		counter("query_agg_avg_rec_count", "query agg avg rec count"),
		counter("query_agg_error", "query agg error"),
		counter("query_agg_success", "query agg success"),
		counter("query_agg", "query agg"),
		counter("query_fail", "query fail"),
		counter("query_long_queue_full", "query long queue full"),
		counter("query_long_reqs", "query long reqs"),
		counter("query_lookup_abort", "query lookup abort"),
		counter("query_lookup_avg_rec_count", "query lookup avg rec count"),
		counter("query_lookup_error", "query lookup error"),
		counter("query_lookup_success", "query lookup success"),
		counter("query_lookups", "query lookups"),
		counter("query_reqs", "query reqs"),
		counter("query_short_queue_full", "query short queue full"),
		counter("query_short_reqs", "query short reqs"),
		counter("query_udf_bg_failure", "query udf bg failure"),
		counter("query_udf_bg_success", "query udf bg success"),
		counter("retransmit_batch_sub_dup_res", "retransmit batch sub dup res"),
		counter("retransmit_client_delete_dup_res", "retransmit client delete dup res"),
		counter("retransmit_client_delete_repl_write", "retransmit client delete repl write"),
		counter("retransmit_client_read_dup_res", "retransmit client read dup res"),
		counter("retransmit_client_udf_dup_res", "retransmit client udf dup res"),
		counter("retransmit_client_udf_repl_write", "retransmit client udf repl write"),
		counter("retransmit_client_write_dup_res", "retransmit client write dup res"),
		counter("retransmit_client_write_repl_write", "retransmit client write repl write"),
		counter("retransmit_nsup_repl_write", "retransmit nsup repl write"),
		counter("retransmit_udf_sub_dup_res", "retransmit udf sub dup res"),
		counter("retransmit_udf_sub_repl_write", "retransmit udf sub repl write"),
		counter("scan_aggr_abort", "scan aggr abort"),
		counter("scan_aggr_complete", "scan aggr complete"),
		counter("scan_aggr_error", "scan aggr error"),
		counter("scan_basic_abort", "scan basic abort"),
		counter("scan_basic_complete", "scan basic complete"),
		counter("scan_basic_error", "scan basic error"),
		counter("scan_udf_bg_abort", "scan udf bg abort"),
		counter("scan_udf_bg_complete", "scan udf bg complete"),
		counter("scan_udf_bg_error", "scan udf bg error"),
		counter("udf_sub_lang_delete_success", "udf sub lang delete success"),
		counter("udf_sub_lang_error", "udf sub lang error"),
		counter("udf_sub_lang_read_success", "udf sub lang read success"),
		counter("udf_sub_lang_write_success", "udf sub lang write success"),
		counter("udf_sub_tsvc_error", "udf sub tsvc error"),
		counter("udf_sub_tsvc_timeout", "udf sub tsvc timeout"),
		counter("udf_sub_udf_complete", "udf sub udf complete"),
		counter("udf_sub_udf_error", "udf sub udf error"),
		counter("udf_sub_udf_timeout", "udf sub udf timeout"),
		counter("xdr_write_error", "xdr write error"),
		counter("xdr_write_success", "xdr write success"),
		counter("xdr_write_timeout", "xdr write timeout"),
		gauge("available_bin_names", "available bin names"),
		gauge("device_available_pct", "device available pct"),
		gauge("device_compression_ratio", "device compression ratio"),
		gauge("device_free_pct", "device free pct"),
		gauge("device_total_bytes", "device total bytes"),
		gauge("device_used_bytes", "device used bytes"),
		gauge("effective_is_quiesced", "effective is quiesced"),
		gauge("effective_replication_factor", "effective replication factor"),
		gauge("evict-hist-buckets", "evict hist buckets"),
		gauge("evict-tenths-pct", "evict tenths pct"),
		gauge("high-water-disk-pct", "high water disk pct"),
		gauge("high-water-memory-pct", "high water memory pct"),
		gauge("hwm_breached", "hwm breached"),
		gauge("index_flash_used_bytes", "index flash used bytes"),
		gauge("index_flash_used_pct", "index flash used pct"),
		gauge("index-type.mounts-high-water-pct", "index type mounts high water pct"),
		gauge("index-type.mounts-size-limit", "index type mounts size limit"),
		gauge("master_objects", "master objects"),
		gauge("master_tombstones", "master tombstones"),
		gauge("memory_free_pct", "memory free pct"),
		gauge("memory_used_bytes", "memory used bytes"),
		gauge("memory_used_data_bytes", "memory used data bytes"),
		gauge("memory_used_index_bytes", "memory used index bytes"),
		gauge("memory_used_sindex_bytes", "memory used sindex bytes"),
		gauge("memory-size", "memory size"),
		gauge("migrate_record_receives", "migrate record receives"),
		gauge("migrate_record_retransmits", "migrate record retransmits"),
		gauge("migrate_records_skipped", "migrate records skipped"),
		gauge("migrate_records_transmitted", "migrate records transmitted"),
		gauge("migrate_rx_instances", "migrate rx instances"),
		gauge("migrate_rx_partitions_active", "migrate rx partitions active"),
		gauge("migrate_rx_partitions_initial", "migrate rx partitions initial"),
		gauge("migrate_rx_partitions_remaining", "migrate rx partitions remaining"),
		gauge("migrate_signals_active", "migrate signals active"),
		gauge("migrate_signals_remaining", "migrate signals remaining"),
		gauge("migrate_tx_instances", "migrate tx instances"),
		gauge("migrate_tx_partitions_active", "migrate tx partitions active"),
		gauge("migrate_tx_partitions_imbalance", "migrate tx partitions imbalance"),
		gauge("migrate_tx_partitions_initial", "migrate tx partitions initial"),
		gauge("migrate_tx_partitions_remaining", "migrate tx partitions remaining"),
		gauge("n_nodes_quiesced", "n nodes quiesced"),
		gauge("non_expirable_objects", "non expirable objects"),
		gauge("non_replica_objects", "non replica objects"),
		gauge("non_replica_tombstones", "non replica tombstones"),
		gauge("ns_cluster_size", "ns cluster size"),
		gauge("objects", "objects"),
		gauge("pending_quiesce", "pending quiesce"),
		gauge("prole_objects", "prole objects"),
		gauge("prole_tombstones", "prole tombstones"),
		gauge("replication-factor", "replication factor"),
		gauge("stop_writes", "stop writes"),
		gauge("stop-writes-pct", "stop writes pct"),
		gauge("tombstones", "tombstones"),
		// including additional key metrics as recommended by aerospike  https://www.aerospike.com/docs/operations/monitor/key_metrics/index.html
		gauge("clock_skew_stop_writes", "clock skew stop writes"),
		gauge("dead_partitions", "dead partitions"),
		gauge("unavailable_partitions", "unavailable partitions"),
		gauge("rack-id", "rack id"),
		// device-level stats don't appear to work
		// and this plugin thinks "storage-engine.device[0].write_q" is malformed.
	}
	NamespaceStorageMetrics = []metric{
		counter("defrag_reads", "defrag reads"),
		counter("defrag_writes", "defrag writes"),
		gauge("shadow_write_q", "shadow write queue"),
		gauge("defrag_q", "defrag queue"),
		gauge("write_q", "write queue"),
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
	for _, m := range NamespaceStorageMetrics {
		ns[m.aeroName] = cmetric{
			typ: m.typ,
			desc: prometheus.NewDesc(
				promkey(systemNamespace, m.aeroName),
				m.desc,
				[]string{"namespace", "mount"},
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

func (nc nsCollector) parseStorage(s string, d string) (string, error) {
	// the function remove the storage prefix metrics for each device:
	// d is storage-engine.device[ix]
	// s is all storage metrics that has been scraped
	// storage-engine.device[ix].age -> age
	// https://www.aerospike.com/docs/reference/metrics/#storage-engine.device[ix].age
	buf := bytes.Buffer{}
	for _, l := range strings.Split(s, ";") {
		for _, v := range strings.Split(l, ":") {
			kv := strings.SplitN(v, "=", 2)
			if len(kv) > 1 {
				if strings.HasPrefix(kv[0], d) {
					//todo: optimize
					kv[0] = strings.Replace(kv[0]+".", d, "", 1)
					kv[0] = strings.Replace(kv[0], ".", "", -1)
				}
				buf.WriteString(kv[0] + "=" + kv[1] + ";")
			}
		}
	}
	r := buf.String()
	return r, nil
}

func (nc nsCollector) splitInfo(s string) (string, string, map[string]string) {
	nsStorageMounts := map[string]string{}

	bufStandardMetrics := bytes.Buffer{}
	bufStorageMetrics := bytes.Buffer{}

	for _, l := range strings.Split(s, ";") {
		for _, v := range strings.Split(l, ":") {
			kv := strings.SplitN(v, "=", 2)
			if strings.HasPrefix(kv[0], "storage-engine") {
				bufStorageMetrics.WriteString(v + ";")
				if strings.HasSuffix(kv[0], "]") {
					nsStorageMounts[kv[1]] = kv[0]
				}
			} else {
				bufStandardMetrics.WriteString(v + ";")
			}
		}
	}
	nsStandardMetrics := bufStandardMetrics.String()
	nsStorageMetrics := bufStorageMetrics.String()
	return nsStorageMetrics, nsStandardMetrics, nsStorageMounts
}

func (nc nsCollector) collect(conn *as.Connection) ([]prometheus.Metric, error) {
	info, err := as.RequestInfo(conn, "namespaces")
	if err != nil {
		return nil, err
	}
	var metrics []prometheus.Metric
	for _, ns := range strings.Split(info["namespaces"], ";") {
		nsInfo, err := as.RequestInfo(conn, "namespace/"+ns)
		if err != nil {
			return nil, err
		}

		nsInfoStorage, nsInfoStandard, nsInfoStorageDevices := nc.splitInfo(nsInfo["namespace/"+ns])

		metrics = append(
			metrics,
			infoCollect(cmetrics(nc), nsInfoStandard, ns)...,
		)

		for mountName, metricName := range nsInfoStorageDevices {
			nsInfoStorage, err = nc.parseStorage(nsInfoStorage, metricName)

			if err != nil {
				return nil, err
			}

			metrics = append(
				metrics,
				infoCollect(cmetrics(nc), nsInfoStorage, ns, mountName)...,
			)
		}
	}
	return metrics, nil
}
