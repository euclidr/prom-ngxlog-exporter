package exporter

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/satyrius/gonx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v2"
)

type MockMetrics struct {
	mock.Mock
}

func (m *MockMetrics) Observe(entry *gonx.Entry, labelValues []string) {
	fmt.Println("hahahahahahah")
	m.Called(entry, labelValues)
}

func (m *MockMetrics) GetLabels() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func TestNamespace(t *testing.T) {

	configText := `
name: api
labels: ["method", "status", "path"]
format: "$remote_addr - $remote_user [$time_local] \"$request\" $status r:$request_length s:$bytes_sent($gzip_ratio) \"$http_referer\" \"$http_user_agent\" ($upstream_addr) $request_time $upstream_response_time $pipe"
default_relabels:
- name: method
  source: request
  split: 1
- name: status
  source: status
- name: path
  source: request
  split: 2
  preprocesses:
  - regexp: "^(.*)\\?.*$"
    replacement: "$1"
  exact_matches:
  - match: "/path/to/1"
    replacement: "/path/to/1"
  - match: "/path/to"
    replacement: "/path/to"
  regex_matches:
  - regexp: "^/user/(.*)$"
    replacement: "/user/<name>"
  - regexp: "^.*$"
    replacement: ""
histogram_buckets: [0.005, 0.01, 0.025, 0.05]
apps:
- name: app1
  source_files:
  - "<placeholder>"
`
	nsCfg := &NamespaceConfig{}
	err := yaml.Unmarshal([]byte(configText), nsCfg)
	assert.NoError(t, err, "unmarshal config yaml failed")

	expectLabels := []string{"app"}
	expectLabels = append(expectLabels, nsCfg.Labels...)

	tests := []struct {
		ngxlog           string
		labels           []string
		values           []string
		requestTimeValue float64
	}{
		{
			ngxlog:           `124.23.132.92 - - [31/Mar/2020:16:35:07 +0800] "POST /api/stat/event HTTP/1.1" 200 r:1297 s:353(-) "-" "-" (127.0.0.1:2300) 0.019 0.016 .`,
			labels:           expectLabels,
			values:           []string{"app1", "POST", "200", ""},
			requestTimeValue: 0.19,
		},
		// {
		// 	ngxlog:           `124.23.132.92 - - [31/Mar/2020:16:35:07 +0800] "POST /api/path/to/1?abc=dee HTTP/1.1" 200 r:1297 s:353(-) "-" "-" (127.0.0.1:2300) 0.019 0.016 .`,
		// 	labels:           expectLabels,
		// 	values:           []string{"app1", "POST", "200", "/path/to/1"},
		// 	requestTimeValue: 0.19,
		// },
		// {
		// 	ngxlog:           `124.23.132.92 - - [31/Mar/2020:16:35:07 +0800] "POST /api/path/to?abc=dee HTTP/1.1" 200 r:1297 s:353(-) "-" "-" (127.0.0.1:2300) 0.019 0.016 .`,
		// 	labels:           expectLabels,
		// 	values:           []string{"app1", "POST", "200", "/path/to"},
		// 	requestTimeValue: 0.19,
		// },
		// {
		// 	ngxlog:           `124.23.132.92 - - [31/Mar/2020:16:35:07 +0800] "POST /user/jack?abc=dee HTTP/1.1" 200 r:1297 s:353(-) "-" "-" (127.0.0.1:2300) 0.019 0.016 .`,
		// 	labels:           expectLabels,
		// 	values:           []string{"app1", "POST", "200", "/user/<name>"},
		// 	requestTimeValue: 0.19,
		// },
	}

	// MatchedBy
	for _, test := range tests {
		// metrics = newMetrics("api", test.labels, nsCfg.HistogramBuckets)
		metrics := MockMetrics{}
		metrics.On("Observe", mock.MatchedBy(
			func(entry *gonx.Entry) bool {
				value, err := entry.FloatField("request_time")
				if err != nil {
					return false
				}
				return value == test.requestTimeValue
			},
		), mock.MatchedBy(
			func(values []string) bool {
				return reflect.DeepEqual(values, test.values)
			},
		)).Return()

		metrics.On("GetLabels").Return(test.labels)

		file, _ := ioutil.TempFile("", "test_exporter")
		nsCfg.Apps[0].SourceFiles[0] = file.Name()

		ns := NewNamespaceWithMetrics(*nsCfg, &metrics)

		stopChan := make(chan bool)

		go ns.StartObserve(stopChan)

		file.Write([]byte(test.ngxlog))
		// fmt.Println("to write", test.ngxlog)
		// fmt.Println("write error", e)
		file.Sync()
		file.Close()

		time.Sleep(time.Millisecond * 5000)

		dx, _ := os.Open(file.Name())
		byts := make([]byte, 0)
		dx.Read(byts)
		fmt.Println("readed", string(byts))

		close(stopChan)

		// os.Remove(file.Name())

		metrics.AssertNumberOfCalls(t, "Observe", 1)
		metrics.AssertNumberOfCalls(t, "GetLabels", 1)
	}
}
