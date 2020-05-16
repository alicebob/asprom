# Project status

Since Aerospike 4.5 (beginning 2020) the Aerospike team have their own Prometheus exporter. It's likely a better choice if you run version 4.5 or higher of Aerospike, especially if you run the Enterprise Edition. The exporter here might be a better choice is you run an older version. PRs are still welcome, but don't expect active maintenance to stay up to date with the current Aerospike.

https://github.com/aerospike/aerospike-prometheus-exporter


# Aerospike Prometheus exporter

This follows the logic from [asgraphite](https://github.com/aerospike/aerospike-graphite). Run a `asprom` collector against every node in the aerospike cluster.

Statistics collected:

  * aerospike_node_*: node wide statistics. e.g. memory usage, cluster state.
  * aerospike_ns_*: per namespace. e.g. objects, migrations.
  * aerospike_sets_*: statistics per set: objects, memory usage
  * aerospike_latency_*: read/write/etc latency rates(!), per namespace
  * aerospike_ops_*: read/write/etc ops per second, per namespace

## Binaries

The [releases](https://github.com/alicebob/asprom/releases) page has binaries.

## Building

- install the [Go compiler](https://golang.org/dl)
- run `make`
- copy the `./asprom` binary to where you need it

It's also easy to crosscompile with Go. You can build asprom for Linux on a Mac with: `GOOS=linux GOARCH=amd64 go build` and then copy the `asprom` binary over to your Linux machines.
