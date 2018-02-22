// Aerospike prometheus exporter
//
// Collects statistics for a single Aerospike node and makes it available as
// metrics for Prometheus.
//
// Statistics collected:
//   aerospike_node_*: node wide statistics. e.g. memory usage, cluster state.
//   aerospike_ns_*: per namespace. e.g. objects, migrations.
//   aerospike_sets_*: statistics per set: objects, memory usage
//   aerospike_latency_*: read/write/etc latency rates(!) (as asinfo -v "latency:" reports").
//   aerospike_ops_*: read/write/etc ops per second
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace       = "aerospike"
	systemNode      = "node"
	systemNamespace = "ns"
	systemLatency   = "latency"
	systemOps       = "ops" // reported in latency
	systemSet       = "set"
)

var (
	version     = "master"
	showVersion = flag.Bool("version", false, "show version")
	addr        = flag.String("listen", ":9145", "listen address for prometheus")
	nodeAddr    = flag.String("node", "127.0.0.1:3000", "aerospike node")
	username    = flag.String("username", "", "username. Leave empty for no authentication")
	password    = flag.String("password", "", "password")

	landingPage = `<html>
<head><title>Aerospike exporter</title></head>
<body>
<h1>Aerospike exporter</h1>
<p><a href="/metrics">Metrics</a></p>
</body>
</html>`

	upDesc = prometheus.NewDesc(
		namespace+"_"+systemNode+"_up",
		"Is this node up",
		nil,
		nil,
	)
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 0 {
		log.Fatal("usage error")
	}

	if *showVersion {
		fmt.Printf("asprom %s\n", version)
		os.Exit(0)
	}

	col := newAsCollector(*nodeAddr, *username, *password)

	req := prometheus.NewRegistry()
	req.MustRegister(col)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(landingPage))
	})
	http.Handle("/metrics", promhttp.HandlerFor(req, promhttp.HandlerOpts{}))
	log.Printf("starting asprom. listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

type collector interface {
	collect(*as.Connection, chan<- prometheus.Metric)
	describe(ch chan<- *prometheus.Desc)
}

type asCollector struct {
	nodeAddr     string
	username     string
	password     string
	totalScrapes prometheus.Counter
	collectors   []collector
}

func newAsCollector(nodeAddr, username, password string) *asCollector {
	totalScrapes := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: systemNode,
		Name:      "scrapes_total",
		Help:      "Total number of times Aerospike was scraped for metrics.",
	})

	return &asCollector{
		nodeAddr:     nodeAddr,
		username:     username,
		password:     password,
		totalScrapes: totalScrapes,
		collectors: []collector{
			newStatsCollector(),
			newNSCollector(),
			newLatencyCollector(),
			newSetCollector(),
		},
	}
}

// Describe implements the prometheus.Collector interface.
func (asc *asCollector) Describe(ch chan<- *prometheus.Desc) {
	asc.totalScrapes.Describe(ch)
	ch <- upDesc
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
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0.0)
		return
	}
	if asc.username != "" {
		if err := conn.Authenticate(asc.username, []byte(asc.password)); err != nil {
			log.Printf("auth error: %s", err)
			ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0.0)
			return
		}
	}
	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 1.0)

	defer conn.Close()

	for _, c := range asc.collectors {
		c.collect(conn, ch)
	}
}
