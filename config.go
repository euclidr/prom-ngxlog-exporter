package main

import "github.com/euclidr/prom-ngxlog-exporter/exporter"


type Config struct {
	Listen     ListenConfig      `yaml:"listen"`
	Namespaces []exporter.NamespaceConfig `yaml:"namespaces"`
}

type ListenConfig struct {
	Port    int    `yaml:"port"`
	Address string `yaml:"address"`
}
