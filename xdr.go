package main

import (
  "strings"

  as "github.com/aerospike/aerospike-client-go"
  "github.com/prometheus/client_golang/prometheus"
)

var (
  // DCMetrics lists the keys we report from aero's dc statistics command.
  // See `asinfo -l -v dcs` for a list of XDR DCs.
  // See `asinfo -l -v dc/<dc>` for detailed metrics for a given DC.
  DCMetrics = []metric{
    gauge("dc_as_open_conn", "Number of open connection to the Aerospike DC."),
    gauge("dc_as_size", "The cluster size of the destination Aerospike DC."),
    gauge("dc_http_good_locations", "Number of URLs that are considered healthy."),
    gauge("dc_http_locations", "Number of URLs configured for the HTTP destination."),
    counter("dc_ship_attempt", "Number of records that have been attempted to be shipped."),
    counter("dc_ship_bytes", "Number of bytes shipped for this DC."),
    counter("dc_ship_delete_success", "Number of delete transactions that have been successfully shipped."),
    counter("dc_ship_destination_error", "Number of errors from the remote cluster(s) while shipping records for this DC."),
    gauge("dc_ship_idle_avg", "Average number of ms of sleep for each record being shipped."),
    gauge("dc_ship_idle_avg_pct", "Representation in percent of total time spent for dc_ship_idle_avg."),
    gauge("dc_ship_inflight_objects", "Number of records that are inflight."),
    gauge("dc_ship_latency_avg", "Moving average of shipping latency for the specific DC."),
    counter("dc_ship_source_error", "Number of client layer errors while shipping records for this DC."),
    counter("dc_ship_success", "Number of records that have been successfully shipped."),
    // dc_state https://www.aerospike.com/docs/reference/metrics/?show-removed=0#dc_state
    gauge("dc_timelag", "Time lag for this specific DC."),
  }
)

type XdrDCCollector cmetrics

func newXdrDCCollector() XdrDCCollector {
  dc := map[string]cmetric {}
  for _, m := range DCMetrics {
    dc[m.aeroName] = cmetric{
      typ: m.typ,
      desc: prometheus.NewDesc(
        promkey(xdrDC, m.aeroName),
        m.desc,
        []string{"dc"},
        nil,
      ),
    }
  }
  return dc
}

func (dcc XdrDCCollector) describe(ch chan<- *prometheus.Desc) {
  for _, s := range dcc {
    ch <- s.desc
  }
}

func (sic XdrDCCollector) collect(conn *as.Connection) ([]prometheus.Metric, error) {
  info, err := as.RequestInfo(conn, "dcs")
  if err != nil {
    return nil, err
  }

  var metrics []prometheus.Metric
  for _, dc := range strings.Split(info["dcs"], ";") {
    dcInfo, err := as.RequestInfo(conn, "dc/"+dc)
    if err != nil {
      return nil, err
    }

    metrics = append(
      metrics,
      infoCollect(
        cmetrics(sic),
        dcInfo["dc/"+dc],
        dc,
      )...,
    )
  }
  return metrics, nil
}
