package exporter

// NamespaceConfig config a group of metric scraper
// metrics in a namespace share:
//   * group of metric names
//   * labels
//   * nginx log format
//   * historgram buckets
// a namespace has default relabel configs, they can be changed by child APP observers
type NamespaceConfig struct {
	Name             string           `yaml:"name"`
	Labels           []string         `yaml:"labels"`
	Format           string           `yaml:"format"`
	DefaultRelabels  []*RelabelConfig `yaml:"default_relabels"`
	HistogramBuckets []float64        `yaml:"historgram_buckets"`

	Apps []AppConfig
}

// AppConfig config an APP observer
// APP observer has a name, which will be labeled in metrics as app={name}
// APP observer can read from multple files
// APP observer can overide namespace's default relabel configs
type AppConfig struct {
	Name        string           `yaml:"name"`
	SourceFiles []string         `yaml:"source_files"`
	Relables    []*RelabelConfig `yaml:"relabels"`
}

// RelabelConfig tells the application how to extract a metric label value
type RelabelConfig struct {
	Name    string                `yaml:"name"`   // label name
	Source  string                `yaml:"source"` // keyword in nginx log format
	Split   int                   `yaml:"split"`  // sometimes we need onyl part of source value, we can split source value and take specific part
	Matches []*RelabelMatchConfig `yaml:"matches"`
}

// RelabelMatchConfig tells how to extract label value from source value
type RelabelMatchConfig struct {
	RegexpString string `yaml:"regexp"`      // pattern to match source value
	Replacement  string `yaml:"replacement"` // replace pattern, like "abc_$1"
	Forward      bool   `yaml:"forward"`     // the processor should forward to the next RelabelMatching after source value matched and replaced
}
