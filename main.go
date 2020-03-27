package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"

	"github.com/euclidr/prom-ngxlog-exporter/exporter"
	"github.com/getsentry/sentry-go"
)

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

	if cfg.Sentry != nil {
		sentry.Init(sentry.ClientOptions{
			Dsn:   cfg.Sentry.Dsn,
			Debug: cfg.Sentry.Debug,
		})
	}

	sigChan := make(chan os.Signal, 1)
	stopChan := make(chan bool)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)

	namespaces := make([]*exporter.Namespace, 0)

	for _, nsCfg := range cfg.Namespaces {
		namespace := exporter.NewNamespace(nsCfg)
		namespaces = append(namespaces, namespace)
	}

	wg := sync.WaitGroup{}

	for _, namespace := range namespaces {
		wg.Add(1)
		go func(ns *exporter.Namespace) {
			ns.StartObserve(stopChan)
			wg.Done()
		}(namespace)
	}

	go func() {
		<-sigChan
		close(stopChan)
		wg.Wait()
		os.Exit(0)
	}()

	http.Handle("/metrics", promhttp.Handler())

	if cfg.EnableShutdownAPI {
		http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
			// shutdown after a while
			go func() {
				time.Sleep(time.Millisecond * 50)
				close(stopChan)
				wg.Wait()
				os.Exit(0)
			}()

			fmt.Fprintf(w, "OK")
		})
	}

	listenAddr := fmt.Sprintf("%s:%d", cfg.Listen.Address, cfg.Listen.Port)
	fmt.Println(listenAddr)
	http.ListenAndServe(listenAddr, nil)
}
