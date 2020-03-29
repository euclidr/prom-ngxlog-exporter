package exporter

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/satyrius/gonx"
	"github.com/stretchr/testify/assert"
)

func TestRegexReplace(t *testing.T) {
	tests := []struct {
		cfg           RegexReplaceConfig
		input         string
		outputMatched bool
		output        string
	}{
		{
			cfg: RegexReplaceConfig{
				RegexpString: "/url/path",
				Replacement:  "/path",
			},
			input:         "/url/path",
			outputMatched: true,
			output:        "/path",
		},
		{
			cfg: RegexReplaceConfig{
				RegexpString: "/url/path",
				Replacement:  "/path",
			},
			input:         "/url/pat",
			outputMatched: false,
			output:        "",
		},
		{
			cfg: RegexReplaceConfig{
				RegexpString: "^/url/(.*)$",
				Replacement:  "/$1",
			},
			input:         "/url/haha",
			outputMatched: true,
			output:        "/haha",
		},
		{
			cfg: RegexReplaceConfig{
				RegexpString: "^/url/(.*)$",
				Replacement:  "/$1",
			},
			input:         "/url1/haha",
			outputMatched: false,
			output:        "",
		},
	}

	for idx, test := range tests {
		replacer := NewRegexReplace(&test.cfg)
		matched, output := replacer.Convert(test.input)
		assert.Equal(t, test.outputMatched, matched, "case %d", idx)
		assert.Equal(t, test.output, output, "case %d", idx)
	}
}

func TestRelabeling(t *testing.T) {

	ngxlogFmt := "$remote_addr - $remote_user [$time_local] \"$request\" $status r:$request_length s:$bytes_sent($gzip_ratio) \"$http_referer\" \"$http_user_agent\" ($upstream_addr) $request_time $upstream_response_time $pipe"

	parser := gonx.NewParser(ngxlogFmt)

	ngxlogGenerator := strings.Replace(ngxlogFmt,
		"\"$request\"",
		"\"%s\"",
		1)

	ngxlogGenerator = strings.Replace(ngxlogGenerator,
		"$status",
		"%s",
		1)

	statusRelabeling := Relabeling{
		Name:   "status",
		Source: "status",
	}

	qmarkStriper := &RegexReplace{
		CompiledRegexp: regexp.MustCompile("^(.*?)\\?.*"),
		Replacement:    "$1",
	}

	userPathReplacer := &RegexReplace{
		CompiledRegexp: regexp.MustCompile("^/user/(.*)$"),
		Replacement:    "/user/<uid>",
	}

	wildcardPathReplacer := &RegexReplace{
		CompiledRegexp: regexp.MustCompile(".*"),
		Replacement:    "other",
	}

	pathRelabeling := Relabeling{
		Name:         "path",
		Source:       "request",
		Split:        2,
		Preprocesses: []*RegexReplace{qmarkStriper},
		RegexMatches: []*RegexReplace{userPathReplacer, wildcardPathReplacer},
		ExactMatches: map[string]string{
			"/dog": "/dog",
			"/cat": "/cat",
		},
	}

	tests := []struct {
		relabeling Relabeling
		status     string
		request    string
		output     string
	}{
		{
			relabeling: statusRelabeling,
			status:     "200",
			request:    "GET /dog HTTP/2.0",
			output:     "200",
		},
		{
			relabeling: pathRelabeling,
			status:     "200",
			request:    "GET /dog HTTP/2.0",
			output:     "/dog",
		},
		{
			relabeling: pathRelabeling,
			status:     "200",
			request:    "GET /cat HTTP/2.0",
			output:     "/cat",
		},
		{
			relabeling: pathRelabeling,
			status:     "200",
			request:    "GET /carrot HTTP/2.0",
			output:     "other",
		},
		{
			relabeling: pathRelabeling,
			status:     "200",
			request:    "GET /user/jack HTTP/2.0",
			output:     "/user/<uid>",
		},
		{
			relabeling: pathRelabeling,
			status:     "200",
			request:    "GET /user/jack/son HTTP/2.0",
			output:     "/user/<uid>",
		},
	}

	for idx, test := range tests {
		ngxlog := fmt.Sprintf(ngxlogGenerator, test.request, test.status)
		entry, _ := parser.ParseString(ngxlog)

		output := test.relabeling.Extract(entry)

		assert.Equal(t, test.output, output, "case %d", idx)
	}
}
