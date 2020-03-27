package main

import "github.com/euclidr/prom-ngxlog-exporter/exporter"

type Config struct {
	Listen            ListenConfig               `yaml:"listen"`
	Sentry            *SentryConfig              `yaml:"sentry"`
	EnableShutdownAPI bool                       `yaml:"enable_shutdown_api"`
	Namespaces        []exporter.NamespaceConfig `yaml:"namespaces"`
}

type ListenConfig struct {
	Port    int    `yaml:"port"`
	Address string `yaml:"address"`
}

type SentryConfig struct {
	Dsn   string `yaml:"dsn"`
	Debug bool   `yaml:"debug"`
}
