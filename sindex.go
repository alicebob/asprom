package main

import (
  "strings"

  as "github.com/aerospike/aerospike-client-go"
  "github.com/prometheus/client_golang/prometheus"
)

var (
  // SindexMetrics lists the keys we report from aero's sindex statistics command.
  // See `asinfo -l -v sindex` for a list of secondary indexes.
  // See `asinfo -l -v sindex/<namespace>` for a list of secondary indexes for a given namespace.
  // See `asinfo -l -v sindex/<namespace>/<sindex_name>` for detailed metrics for a given secondary index.
  SindexMetrics = []metric{
    gauge("keys", "keys"),
    gauge("entries", "entries"),
    gauge("ibtr_memory_used", "ibtr_memory_used"),
    gauge("nbtr_memory_used", "nbtr_memory_used"),
    gauge("si_accounted_memory", "si_accounted_memory"),
    gauge("load_pct", "load_pct"),
    counter("loadtime", "loadtime"),
    counter("write_success", "write_success"),
    counter("write_error", "write_error"),
    counter("delete_success", "delete_success"),
    counter("delete_error", "delete_error"),
    counter("stat_gc_recs", "stat_gc_recs"),
    counter("stat_gc_time", "stat_gc_time"),
    counter("query_reqs", "query_reqs"),
    gauge("query_avg_rec_count", "query_avg_rec_count"),
    gauge("query_avg_record_size", "query_avg_record_size"),
    counter("query_agg", "query_agg"),
    gauge("query_agg_avg_rec_count", "query_agg_avg_rec_count"),
    gauge("query_agg_avg_record_size", "query_agg_avg_record_size"),
    counter("query_lookups", "query_lookups"),
    gauge("query_lookup_avg_rec_count", "query_lookup_avg_rec_count"),
    gauge("query_lookup_avg_record_size", "query_lookup_avg_record_size"),
  }
)

type sindexCollector cmetrics

func newSindexCollector() sindexCollector {
  sindex := map[string]cmetric {}
  for _, m := range SindexMetrics {
    sindex[m.aeroName] = cmetric{
      typ: m.typ,
      desc: prometheus.NewDesc(
        promkey(secondaryIndex, m.aeroName),
        m.desc,
        []string{"namespace", "sindex", "set", "bin", "type", "indextype", "path"},
        nil,
      ),
    }
  }
  return sindex
}

func (sindexc sindexCollector) describe(ch chan<- *prometheus.Desc) {
  for _, s := range sindexc {
    ch <- s.desc
  }
}

func (sic sindexCollector) collect(conn *as.Connection) ([]prometheus.Metric, error) {
  info, err := as.RequestInfo(conn, "sindex")
  if err != nil {
    return nil, err
  }

  var metrics []prometheus.Metric
  for _, sindexInfo := range strings.Split(info["sindex"], ";") {
    if sindexInfo == "" {
      continue
    }
    sindexStats := parseInfo(sindexInfo)
    ns := sindexStats["ns"]
    sindexName := sindexStats["indexname"]
    sindexDetails, err := as.RequestInfo(conn, "sindex/"+ns+"/"+sindexName)
    if err != nil {
      return nil, err
    }

    metrics = append(
      metrics,
      infoCollect(
        cmetrics(sic),
        sindexDetails["sindex/"+ns+"/"+sindexName],
        ns,
        sindexName,
        sindexStats["set"],
        sindexStats["bin"],
        sindexStats["type"],
        sindexStats["indextype"],
        sindexStats["path"],
      )...,
    )
  }
  return metrics, nil
}
