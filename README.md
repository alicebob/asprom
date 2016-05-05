Aerospike Prometheus exporter

This follows the logic from [asgraphite](https://github.com/aerospike/aerospike-graphite).

TODOs:

  * only some metrics are currently exported, but the rest could be added easily
  * latency as a proper histogram. The aerospike logfiles have the data: [asloglatency](http://www.aerospike.com/docs/operations/monitor/latency/)
