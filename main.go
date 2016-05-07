// Aerospike prometheus exporter
//
// Collects statistics for a single Aerospike node and makes it available as
// metrics for Prometheus.
//
// Statistics collected:
//   aerospike_node_*: node wide statistics. e.g. memory usage, cluster state.
//   aerospike_ns_*: per namespace. e.g. objects, migrations.
//   aerospike_latency_*: read/write/etc latency rates.
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace       = "aerospike"
	systemNode      = "node"
	systemNamespace = "ns"
	systemLatency   = "latency"
)

var (
	addr     = flag.String("listen", ":9191", "listen address for prometheus")
	nodeAddr = flag.String("node", "127.0.0.1:3000", "aerospike node")

	landingPage = `<html>
<head><title>Aerospike exporter</title></head>
<body>
<h1>Aerospike exporter</h1>
<p><a href="/metrics">Metrics</a></p>
</body>
</html>`
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 0 {
		log.Fatal("usage error")
	}

	col := newAsCollector(*nodeAddr)
	prometheus.MustRegister(col)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(landingPage))
	})
	http.Handle("/metrics", prometheus.Handler())
	log.Printf("starting asprom. listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

type collector interface {
	collect(*as.Connection, chan<- prometheus.Metric)
	describe(ch chan<- *prometheus.Desc)
}

type asCollector struct {
	nodeAddr         string
	totalScrapes, up prometheus.Counter
	collectors       []collector
}

func newAsCollector(nodeAddr string) *asCollector {
	totalScrapes := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: systemNode,
		Name:      "scrapes_total",
		Help:      "Total number of times Aerospike was scraped for metrics.",
	})

	up := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: systemNode,
		Name:      "up",
		Help:      "Is this node up",
	})

	return &asCollector{
		nodeAddr:     nodeAddr,
		totalScrapes: totalScrapes,
		up:           up,
		collectors: []collector{
			newStatsCollector(),
			newNSCollector(),
			newLatencyCollector(),
		},
	}
}

// Describe implements the prometheus.Collector interface.
func (asc *asCollector) Describe(ch chan<- *prometheus.Desc) {
	asc.totalScrapes.Describe(ch)
	asc.up.Describe(ch)
	for _, c := range asc.collectors {
		c.describe(ch)
	}
}

// Collect implements the prometheus.Collector interface.
func (asc *asCollector) Collect(ch chan<- prometheus.Metric) {
	asc.totalScrapes.Inc()
	ch <- asc.totalScrapes

	conn, err := as.NewConnection(asc.nodeAddr, 3*time.Second)
	if err != nil {
		asc.up.Set(0.0)
		ch <- asc.up
		return
	}
	asc.up.Set(1.0)
	ch <- asc.up

	defer conn.Close()

	for _, c := range asc.collectors {
		c.collect(conn, ch)
	}
}
