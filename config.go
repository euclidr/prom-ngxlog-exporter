package main

import "github.com/euclidr/prom-ngxlog-exporter/exporter"

type Config struct {
	Listen     ListenConfig               `yaml:"listen"`
	Sentry     *SentryConfig              `yaml:"sentry"`
	Namespaces []exporter.NamespaceConfig `yaml:"namespaces"`
}

type ListenConfig struct {
	Port    int    `yaml:"port"`
	Address string `yaml:"address"`
}

type SentryConfig struct {
	Dsn   string `yaml:"dsn"`
	Debug bool   `yaml:"debug"`
}
