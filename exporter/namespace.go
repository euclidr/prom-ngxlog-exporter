package exporter

import (
	"sync"

	"github.com/satyrius/gonx"
)

// Namespace collect a group of metrics of different APP
type Namespace struct {
	Name    string
	Parser  *gonx.Parser
	Metrics Metrics

	Labels   []string
	Relabels map[string]*Relabeling

	Observers []*AppObserver

	LabelIndexer map[string]int
}

// NewNamespace create a Namespace object from nsCfg
func NewNamespace(nsCfg NamespaceConfig) *Namespace {
	labels := []string{"app"}
	labels = append(labels, nsCfg.Labels...)
	metrics := newMetrics(nsCfg.Name, labels, nsCfg.HistogramBuckets)

	return NewNamespaceWithMetrics(nsCfg, metrics)
}

// NewNamespaceWithMetrics create a Namespace object from config and metrics
//   It uses labels in metrics and ignore labels in nsCfg
func NewNamespaceWithMetrics(nsCfg NamespaceConfig, metrics Metrics) *Namespace {
	ns := &Namespace{}

	ns.Name = nsCfg.Name
	ns.Parser = gonx.NewParser(nsCfg.Format)
	ns.Metrics = metrics

	ns.Labels = metrics.GetLabels()
	ns.genLabelIndexer()

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

// StartObserve start goroutines to observe metrics of different APP Observers
func (ns *Namespace) StartObserve(stopChan <-chan bool) {
	wg := sync.WaitGroup{}
	for _, observer := range ns.Observers {
		wg.Add(1)
		go func(obs *AppObserver) {
			obs.startObserve(stopChan)
			wg.Done()
		}(observer)
	}
	wg.Wait()
}
