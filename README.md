Aerospike Prometheus exporter

This follows the logic from [asgraphite](https://github.com/aerospike/aerospike-graphite). Run a `asprom` collector against every node in the aerospike cluster.

Statistics collected:

  * aerospike_node_*: node wide statistics. e.g. memory usage, cluster state.
  * aerospike_ns_*: per namespace. e.g. objects, migrations.
  * aerospike_sets_*: statistics per set: objects, memory usage
  * aerospike_latency_*: read/write/etc latency rates(!) (as asinfo -v "latency:" reports").

TODOs:

  * latency as a proper histogram. The server doesn't expose this data, but the aerospike logfiles do have the data: [asloglatency](http://www.aerospike.com/docs/operations/monitor/latency/)
