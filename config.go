package main

type Config struct {
	Listen     ListenConfig      `yaml:"listen"`
	Namespaces []NamespaceConfig `yaml:"namespaces"`
}

type ListenConfig struct {
	Port    int    `yaml:"port"`
	Address string `yaml:"address"`
}

type NamespaceConfig struct {
	Name             string           `yaml:"name"`
	Labels           []string         `yaml:"labels"`
	Format           string           `yaml:"format"`
	DefaultRelabels  []*RelabelConfig `yaml:"default_relabels"`
	HistogramBuckets []float64        `yaml:"historgram_buckets"`

	Apps []AppConfig
}

type AppConfig struct {
	Name        string           `yaml:"name"`
	SourceFiles []string         `yaml:"source_files"`
	Relables    []*RelabelConfig `yaml:"relabels"`
}

type RelabelConfig struct {
	Name    string                `yaml:"name"`
	Source  string                `yaml:"source"`
	Split   int                   `yaml:"split"`
	Matches []*RelabelMatchConfig `yaml:"matches"`
}

type RelabelMatchConfig struct {
	RegexpString string `yaml:"regexp"`
	Replacement  string `yaml:"replacement"`
	Forward      bool   `yaml:"forward"`
}
