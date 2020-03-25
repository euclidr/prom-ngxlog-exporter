package exporter

import (
	"sync"
)

// AppObserver reads incoming nginx logs and gather values
type AppObserver struct {
	Name string

	ns *Namespace

	followers []Follower

	relabels map[string]*Relabeling
}

// NewAppObserver create a new AppObserver
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
	if err != nil {
		// TODO error
		return
	}
	labelValues := make([]string, len(ao.ns.Labels))
	labelValues[ao.ns.LabelIndexer["app"]] = ao.Name

	for name, relabeling := range ao.relabels {
		idx := ao.ns.LabelIndexer[name]
		labelValues[idx] = relabeling.Extract(entry)
	}

	metrics := ao.ns.Metrics
	metrics.Observe(entry, labelValues)
}

func (ao *AppObserver) startObserve(stopChan <-chan bool) {
	wg := sync.WaitGroup{}
	for _, follower := range ao.followers {
		wg.Add(1)
		go func(f Follower) {
			ao.processSource(f, stopChan)
			wg.Done()
		}(follower)
	}
	wg.Wait()
}

func (ao *AppObserver) processSource(follower Follower, stopChan <-chan bool) {
	lineChan := follower.Lines()

Loop:
	for {
		select {
		case line := <-lineChan:
			ao.observeLine(line)
		case <-stopChan:
			break Loop
		}
	}
}
