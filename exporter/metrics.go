package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/satyrius/gonx"
)

// Metrics is an interface to observe value
type Metrics interface {
	Observe(entry *gonx.Entry, labelValues []string)
	GetLabels() []string
}

type defaultMetrics struct {
	labels     []string
	countTotal *prometheus.CounterVec
	bytesTotal *prometheus.CounterVec
	latencies  *prometheus.HistogramVec
}

func newMetrics(namespace string, labels []string, buckets []float64) *defaultMetrics {
	m := &defaultMetrics{}

	m.labels = labels

	m.countTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_count",
		Help:      "Amount of processed HTTP requests",
	}, labels)

	m.bytesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_bytes",
		Help:      "Total amount of transfered bytes",
	}, labels)

	m.latencies = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_requests_latencies",
		Buckets:   buckets,
		Help:      "HTTP request letancies in seconds",
	}, labels)

	prometheus.MustRegister(m.countTotal)
	prometheus.MustRegister(m.bytesTotal)
	prometheus.MustRegister(m.latencies)

	return m
}

func (dm *defaultMetrics) Observe(entry *gonx.Entry, labelValues []string) {
	dm.countTotal.WithLabelValues(labelValues...).Inc()

	if bytes, err := entry.FloatField("bytes_sent"); err == nil {
		dm.bytesTotal.WithLabelValues(labelValues...).Add(bytes)
	}

	if latency, err := entry.FloatField("request_time"); err == nil {
		dm.latencies.WithLabelValues(labelValues...).Observe(latency)
	}
}

func (dm *defaultMetrics) GetLabels() []string {
	return dm.labels
}
