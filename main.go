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
	"crypto/tls"
	"flag"
	"fmt"
	as "github.com/aerospike/aerospike-client-go"
	"github.com/aerospike/aerospike-client-go/pkg/bcrypt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	namespace         = "aerospike"
	secondaryIndex    = "sindex"
	systemNode        = "node"
	systemNamespace   = "ns"
	systemLatency     = "latency"
	systemLatencyHist = "latency_hist" // total number of ops
	systemOps         = "ops"
	systemSet         = "set"
	xdrDC             = "xdr"
)

var (
	version     = "master"
	showVersion = flag.Bool("version", false, "show version")
	addr        = flag.String("listen", ":9145", "listen address for prometheus. ENV variable EXPORTER_ADDRESS")
	nodeAddr    = flag.String("node", "127.0.0.1:3000", "aerospike node")
	tlsName     = flag.String("tlsName", "", "tlsName")
	tlsKey      = flag.String("tlsKey", "", "certificate - key")
	tlsCert     = flag.String("tlsCert", "", "certificate - cert")
	enableTLS   = flag.Bool("enableTLS", false, "enable or disable tls")
	username    = flag.String("username", "", "username. Leave empty for no authentication. ENV variable AS_USERNAME, if set, will override this.")
	password    = flag.String("password", "", "password. ENV variable AS_PASSWORD, if set, will override this.")
	//authMode = flag.String("A", "internal", "Authentication mode: internal | external")

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

func configureClientPolicy(clientPolicy *as.ClientPolicy, username string, password string, certificate string, key string) {

	if username != "" {
		clientPolicy.User = username
		clientPolicy.Password = password
	}
	/*
		if *authMode == "external" {
			clientPolicy.AuthMode = as.AuthModeExternal

		}
	*/
	cert, err := tls.LoadX509KeyPair(certificate, key)
	if err != nil {
		log.Fatal("cert error")
	}

	config := tls.Config{
		Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	clientPolicy.TlsConfig = &config

}

func main() {
	flag.Parse()
	if len(flag.Args()) != 0 {
		log.Fatal("usage error")
	}

	user := os.Getenv("AS_USERNAME")
	if user != "" {
		*username = user
	}

	pass := os.Getenv("AS_PASSWORD")
	if pass != "" {
		*password = pass
	}

    exporterAddr := os.Getenv("EXPORTER_ADDRESS")
    if exporterAddr != "" {
        *addr = exporterAddr
    }

	if *showVersion {
		fmt.Printf("asprom %s\n", version)
		os.Exit(0)
	}
	var port string
	_, port, _ = net.SplitHostPort(*nodeAddr)
	var col *asCollector
	clientPolicy := as.NewClientPolicy()

	if *enableTLS == true {
		if *tlsName == "" || *tlsCert == "" || *tlsKey == "" {
			log.Fatal("You are missing either tlsName, certificate or key for secure connection")
		}
		configureClientPolicy(clientPolicy, *username, *password, *tlsCert, *tlsKey)
		col = newAsCollector(*nodeAddr, *clientPolicy, port, clientPolicy.User, clientPolicy.Password)

	} else {
		//port = 3000 //set default port
		col = newAsCollector(*nodeAddr, *clientPolicy, port, *username, *password)
	}

	req := prometheus.NewRegistry()
	req.MustRegister(col)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(landingPage))
	})
	http.Handle("/metrics", promhttp.HandlerFor(req, promhttp.HandlerOpts{ErrorLog: log.New(os.Stdout, "err: ", 0)}))
	log.Printf("starting asprom. listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

type collector interface {
	collect(*as.Connection) ([]prometheus.Metric, error)
	describe(ch chan<- *prometheus.Desc)
}

type asCollector struct {
	nodeAddr     string
	port         string
	username     string
	password     string
	clientPolicy *as.ClientPolicy
	totalScrapes prometheus.Counter
	collectors   []collector
}

func newAsCollector(nodeAddr string, clientPolicy as.ClientPolicy, port string, username string, password string) *asCollector {
	totalScrapes := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: systemNode,
		Name:      "scrapes_total",
		Help:      "Total number of times Aerospike was scraped for metrics.",
	})

	return &asCollector{
		nodeAddr:     nodeAddr,
		port:         port,
		username:     username,
		password:     password,
		totalScrapes: totalScrapes,
		clientPolicy: &clientPolicy,
		collectors: []collector{
			newLatencyCollector(),
			newNSCollector(),
			newSetCollector(),
			newSindexCollector(),
			newStatsCollector(),
			newXdrDCCollector(),
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

	ms, err := asc.collect()
	if err != nil {
		log.Print(err)
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0.0)
		return
	}
	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 1.0)
	for _, m := range ms {
		ch <- m
	}
}

func (asc *asCollector) collect() ([]prometheus.Metric, error) {
	//clientPolicy = as.NewClientPolicy()
	portAsInt, _ := strconv.Atoi(asc.port)
	host := as.NewHost(asc.nodeAddr, portAsInt)
	host.TLSName = *tlsName
	var conn *as.Connection
	var err error

	if *enableTLS == true {
		conn, err = as.NewSecureConnection(asc.clientPolicy, host) //, 3*time.Second)
	} else {
		conn, err = as.NewConnection(asc.nodeAddr, 3*time.Second)
	}

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if asc.clientPolicy.User != "" {
		hp, err := hashPassword(asc.clientPolicy.Password)
		if err != nil {
			return nil, fmt.Errorf("hashPassword: %s", err)
		}
		if err := conn.Authenticate(asc.clientPolicy.User, hp); err != nil {
			return nil, fmt.Errorf("auth error: %s", err)
		}
	}

	var metrics []prometheus.Metric
	for _, c := range asc.collectors {
		ms, err := c.collect(conn)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, ms...)
	}
	return metrics, nil
}

// take from github.com/aerospike/aerospike-client-go/admin_command.go
func hashPassword(password string) ([]byte, error) {
	// Hashing the password with the cost of 10, with a static salt
	const salt = "$2a$10$7EqJtq98hPqEX7fNZaFWoO"
	hashedPassword, err := bcrypt.Hash(password, salt)
	if err != nil {
		return nil, err
	}
	return []byte(hashedPassword), nil
}
