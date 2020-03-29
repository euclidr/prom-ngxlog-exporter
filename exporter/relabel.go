package exporter

import (
	"regexp"
	"strings"

	"github.com/satyrius/gonx"
)

// RegexReplace match and convert value
type RegexReplace struct {
	CompiledRegexp *regexp.Regexp
	Replacement    string
}

// NewRegexReplace create RegexReplace object
func NewRegexReplace(cfg *RegexReplaceConfig) *RegexReplace {
	rr := &RegexReplace{}

	r, err := regexp.Compile(cfg.RegexpString)
	if err != nil {
		panic(err)
	}

	rr.CompiledRegexp = r
	rr.Replacement = cfg.Replacement

	return rr
}

// Convert convert source value if matched
func (rr *RegexReplace) Convert(before string) (matched bool, after string) {
	if rr.CompiledRegexp.MatchString(before) {
		after = rr.CompiledRegexp.ReplaceAllString(before, rr.Replacement)
		return true, after
	}

	return false, ""
}

// Relabeling extract label value from source value
type Relabeling struct {
	Name         string
	Source       string
	Split        int
	Preprocesses []*RegexReplace
	RegexMatches []*RegexReplace
	ExactMatches map[string]string
}

// NewRelabeling create Relabeling object from config
func NewRelabeling(cfg *RelabelConfig) *Relabeling {
	r := &Relabeling{}
	r.Name = cfg.Name
	r.Source = cfg.Source
	r.Split = cfg.Split

	if cfg.Preprocesses != nil {
		processes := make([]*RegexReplace, 0)
		for _, regexMatchCfg := range cfg.Preprocesses {
			process := NewRegexReplace(regexMatchCfg)
			processes = append(processes, process)
		}
		r.Preprocesses = processes
	}

	if cfg.ExcactMatches != nil {
		r.ExactMatches = make(map[string]string)
		for _, exactMatchCfg := range cfg.ExcactMatches {
			r.ExactMatches[exactMatchCfg.Match] = exactMatchCfg.Replacement
		}
	}

	if cfg.RegexMatches != nil {
		matches := make([]*RegexReplace, 0)
		for _, regexMatchCfg := range cfg.RegexMatches {
			match := NewRegexReplace(regexMatchCfg)
			matches = append(matches, match)
		}
		r.RegexMatches = matches
	}

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

	if r.Preprocesses != nil {
		for _, process := range r.Preprocesses {
			matched, after := process.Convert(sourceValue)
			if matched {
				sourceValue = after
				break
			}
		}
	}

	if r.ExactMatches != nil {
		after, matched := r.ExactMatches[sourceValue]
		if matched {
			return after
		}
	}

	if r.RegexMatches != nil {
		for _, matcher := range r.RegexMatches {
			matched, after := matcher.Convert(sourceValue)
			if matched {
				return after
			}
		}
	}

	return sourceValue
}
