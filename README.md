# prom-ngxlog-exporter

Export request metrics from Nginx access log

### Build & Run

clone the project then build:

```
go build -o prom-ngxlog-exporter
```

write a config file then run:

```
./ prom-ngxlog-exporter -config-file /path/to/config.yml
```

### Configuration

#### Concepts

* Namespace: a group of metric, take `name` as metric name prefix
* Label: label of metrics, in a namespace, the set of labels is fixed.
* Nginx Log Format: `log_format` defined in Nginx config file, see [nginx docs](https://docs.nginx.com/nginx/admin-guide/monitoring/logging/#setting-up-the-access-log) for detail
* Relabel: convert value from nginx log to value that set to `Label`, It may need several step to extract that value.
* HistogramBuckets: see [prometheus docs](https://prometheus.io/docs/concepts/metric_types/#histogram) for detail
* App: Application tha share a namespace of metrics, it tells exporter where to find the accesslog files and it can have its specific `Relabel` config.

Please read config schema from `config.go` for unmetioned config params, it's pretty straight forward.

