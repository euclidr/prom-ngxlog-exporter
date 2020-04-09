# prom-ngxlog-exporter

[![Build Status](https://travis-ci.com/euclidr/prom-ngxlog-exporter.svg?branch=master)](https://travis-ci.com/euclidr/prom-ngxlog-exporter)
[![codecov](https://codecov.io/gh/euclidr/prom-ngxlog-exporter/branch/master/graph/badge.svg)](https://codecov.io/gh/euclidr/prom-ngxlog-exporter)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/euclidr/prom-ngxlog-exporter)

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

### Example

```
listen:
  port: 4040
  address: 0.0.0.0

sentry:
  dsn: ""
  debug: true

namespaces:
  - name: api
    labels:
      - method
      - status
      - path

    format: "$remote_addr - $remote_user [$time_local] \"$request\" $status r:$request_length s:$bytes_sent($gzip_ratio) \"$http_referer\" \"$http_user_agent\" ($upstream_addr) $request_time $upstream_response_time $pipe"

    default_relabels:
      - name: method
        source: request
        split: 1
      - name: status
        source: status
      - name: path
        source: request
        split: 2
        preprocesses:
          - regexp: "^(.*)\\?.*$"
            replacement: "$1"
        exact_matches:
          - match: "/api/happen"
            replacement: "/api/happen"
        regex_matches:
          - regexp: "^/static/(.*)$"
            replacement: "/static/<filename>"
          - regexp: "^.*$"
            replacement: ""

    histogram_buckets: [.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10]

    apps:
      - name: app1
        source_files:
          - "./dev_data/access.log"
```

#### Concepts

* Namespace: a group of metric, take `name` as metric name prefix
* Label: label of metrics, in a namespace, the set of labels is fixed.
* Nginx Log Format: `log_format` defined in Nginx config file, see [nginx docs](https://docs.nginx.com/nginx/admin-guide/monitoring/logging/#setting-up-the-access-log) for detail
* Relabel: convert value from nginx log to value that set to `Label`, It may need several step to extract that value.
* HistogramBuckets: see [prometheus docs](https://prometheus.io/docs/concepts/metric_types/#histogram) for detail
* App: Application tha share a namespace of metrics, it tells exporter where to find the accesslog files and it can have its specific `Relabel` config.

#### Relabel Procedure

```
+-------------------------+
|      source value       | -+
+-------------------------+  |
  |                          |
  | split > 0                |
  v                          |
+-------------------------+  |
|     split by space      |  |
|    get the Nth part     |  | split = 0
+-------------------------+  |
  |                          |
  |                          |
  v                          |
+-------------------------+  |
|       preprocess        | <+
+-------------------------+
  |
  | result
  v
+-------------------------+
| exact match and replace | -+
+-------------------------+  |
  |                          |
  | not matched              |
  v                          |
+-------------------------+  |
| regex match and replace |  | matched
+-------------------------+  |
  |                          |
  |                          |
  v                          |
+-------------------------+  |
|      got the value      | <+
+-------------------------+
```

Please read config schema from `config.go` for unmetioned config params, it's pretty straight forward.

