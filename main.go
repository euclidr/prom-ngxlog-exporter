package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/satyrius/gonx"
	"gopkg.in/yaml.v2"
)

type Metrics struct {
	countTotal *prometheus.CounterVec
	bytesTotal *prometheus.CounterVec
	latencies  *prometheus.HistogramVec
}

func NewMetrics(namespace string, labels []string, buckets []float64) *Metrics {
	m := &Metrics{}

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

type Namespace struct {
	Name    string
	Parser  *gonx.Parser
	Metrics *Metrics

	Labels   []string
	Relabels map[string]*Relabeling

	Observers []*AppObserver

	LabelIndexer map[string]int
}

func NewNamespace(nsCfg NamespaceConfig) *Namespace {

	ns := &Namespace{}

	ns.Name = nsCfg.Name
	ns.Parser = gonx.NewParser(nsCfg.Format)

	labels := []string{"app"}
	labels = append(labels, nsCfg.Labels...)
	ns.Labels = labels

	ns.genLabelIndexer()

	ns.Metrics = NewMetrics(nsCfg.Name, labels, nsCfg.HistogramBuckets)

	ns.Relabels = make(map[string]*Relabeling)
	for _, relabelCfg := range nsCfg.DefaultRelabels {
		relabeling := NewRelabeling((relabelCfg))
		ns.Relabels[relabeling.Name] = relabeling
	}

	observers := make([]*AppObserver, 0)

	for _, appCfg := range nsCfg.Apps {
		observer := NewAppObserver(ns, appCfg)
		observers = append(observers, observer)
	}

	ns.Observers = observers

	return ns
}

func (ns *Namespace) genLabelIndexer() {
	ns.LabelIndexer = make(map[string]int)
	for idx, label := range ns.Labels {
		ns.LabelIndexer[label] = idx
	}
}

func (ns *Namespace) startObserve() {
	for _, observer := range ns.Observers {
		go observer.startObserve()
	}
}

type AppObserver struct {
	ns *Namespace

	followers []Follower

	Name     string
	relabels map[string]*Relabeling
}

func NewAppObserver(ns *Namespace, appCfg AppConfig) *AppObserver {
	ao := &AppObserver{}

	ao.Name = appCfg.Name

	ao.ns = ns
	followers := make([]Follower, 0)
	for _, filepath := range appCfg.SourceFiles {
		t, err := NewFileFollower(filepath)
		if err != nil {
			panic(err)
		}

		t.OnError(func(err error) {
			panic(err)
		})

		followers = append(followers, t)
	}

	ao.followers = followers

	ao.relabels = make(map[string]*Relabeling)
	for name, relabeling := range ns.Relabels {
		ao.relabels[name] = relabeling
	}

	if appCfg.Relables != nil {
		for _, relabelCfg := range appCfg.Relables {
			relabeling := NewRelabeling((relabelCfg))
			ao.relabels[relabeling.Name] = relabeling
		}
	}

	return ao
}

func (ao *AppObserver) observeLine(line string) {
	parser := ao.ns.Parser
	entry, err := parser.ParseString(line)
	// fmt.Println("entry", entry, err)
	if err != nil {
		// TODO error
		return
	}
	labelValues := make([]string, len(ao.ns.Labels))
	labelValues[ao.ns.LabelIndexer["app"]] = ao.Name

	fmt.Println("ao.Name", ao.Name)

	for name, relabeling := range ao.relabels {
		idx := ao.ns.LabelIndexer[name]
		labelValues[idx] = relabeling.Extract(entry)
	}

	fmt.Println("labelValues", labelValues)

	metrics := ao.ns.Metrics

	metrics.countTotal.WithLabelValues(labelValues...).Inc()

	if bytes, err := entry.FloatField("bytes_sent"); err == nil {
		metrics.bytesTotal.WithLabelValues(labelValues...).Add(bytes)
	}

	if latency, err := entry.FloatField("request_time"); err == nil {
		metrics.latencies.WithLabelValues(labelValues...).Observe(latency)
	}
}

func (ao *AppObserver) startObserve() {
	for _, follower := range ao.followers {
		go ao.processSource(follower)
	}
}

func (ao *AppObserver) processSource(follower Follower) {
	for line := range follower.Lines() {
		ao.observeLine(line)
	}
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "config-file", "", "Configuration file to read from")
	flag.Parse()

	confData, err := ioutil.ReadFile((configFile))
	if err != nil {
		panic(err)
	}

	var cfg Config

	err = yaml.Unmarshal([]byte(confData), &cfg)
	if err != nil {
		panic(err)
	}

	namespaces := make([]*Namespace, 0)

	for _, nsCfg := range cfg.Namespaces {
		namespace := NewNamespace(nsCfg)
		namespaces = append(namespaces, namespace)
	}

	for _, namespace := range namespaces {
		namespace.startObserve()
	}

	http.Handle("/metrics", promhttp.Handler())

	listenAddr := fmt.Sprintf("%s:%d", cfg.Listen.Address, cfg.Listen.Port)
	fmt.Println(listenAddr)
	http.ListenAndServe(listenAddr, nil)
}
