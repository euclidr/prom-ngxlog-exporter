package exporter

import (
	"regexp"
	"strings"

	"github.com/satyrius/gonx"
)

// RelabelMatching match and convert value
type RelabelMatching struct {
	CompiledRegexp *regexp.Regexp
	Replacement    string
	Forward        bool
}

// NewRelabelMatching create RelabelMatching object from config
func NewRelabelMatching(cfg *RelabelMatchConfig) *RelabelMatching {
	m := &RelabelMatching{}

	r, err := regexp.Compile(cfg.RegexpString)
	if err != nil {
		panic(err)
	}

	m.CompiledRegexp = r
	m.Replacement = cfg.Replacement
	m.Forward = cfg.Forward

	return m
}

// Convert convert source value if matched
func (rm *RelabelMatching) Convert(before string) (matched bool, after string) {
	if rm.CompiledRegexp.MatchString(before) {
		after = rm.CompiledRegexp.ReplaceAllString(before, rm.Replacement)
		return true, after
	}

	return false, ""
}

// Relabeling extract label value from source value
type Relabeling struct {
	Name    string
	Source  string
	Split   int
	Matches []*RelabelMatching
}

// NewRelabeling create Relabeling object from config
func NewRelabeling(cfg *RelabelConfig) *Relabeling {
	r := &Relabeling{}
	r.Name = cfg.Name
	r.Source = cfg.Source
	r.Split = cfg.Split

	matches := make([]*RelabelMatching, 0)

	if cfg.Matches != nil {
		for _, matchCfg := range cfg.Matches {
			match := NewRelabelMatching(matchCfg)
			matches = append(matches, match)
		}
	}

	r.Matches = matches

	return r
}

// Extract extract label value from nginx log entry
func (r *Relabeling) Extract(entry *gonx.Entry) string {
	sourceValue, err := entry.Field((r.Source))
	if err != nil {
		return ""
	}

	if r.Split > 0 {
		values := strings.Split(sourceValue, " ")

		if len(values) >= r.Split {
			sourceValue = values[r.Split-1]
		} else {
			sourceValue = ""
		}
	}

	for _, match := range r.Matches {
		matched, after := match.Convert(sourceValue)
		if !matched {
			continue
		}

		sourceValue = after

		if !match.Forward {
			break
		}
	}

	return sourceValue
}
