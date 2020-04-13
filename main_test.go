package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

var message = []byte(`
{
  "receiver": "Default",
  "status": "firing",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "CPUThrottlingHigh",
        "container": "config-reloader",
        "namespace": "prometheus",
        "pod": "alertmanager-main-0",
        "prometheus": "prometheus/prometheus",
        "severity": "warning"
      },
      "annotations": {
        "message": "33.37% throttling of CPU in namespace prometheus for container config-reloader in pod alertmanager-main-0.",
        "runbook_url": "https://github.com/kubernetes-monitoring/kubernetes-mixin/tree/master/runbook.md#alert-name-cputhrottlinghigh"
      },
      "startsAt": "2021-04-10T03:31:01.534463406Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://prometheus-prometheus-1:9090/graph?g0.expr=sum+by%28container%2C+pod%2C+namespace%29+%28increase%28container_cpu_cfs_throttled_periods_total%7Bcontainer%21%3D%22%22%7D%5B5m%5D%29%29+%2F+sum+by%28container%2C+pod%2C+namespace%29+%28increase%28container_cpu_cfs_periods_total%5B5m%5D%29%29+%3E+%2825+%2F+100%29&g0.tab=1",
      "fingerprint": "c9ef92cd8b0e0db3"
    },
    {
      "status": "firing",
      "labels": {
        "alertname": "CPUThrottlingHigh",
        "container": "config-reloader",
        "namespace": "prometheus",
        "pod": "alertmanager-main-1",
        "prometheus": "prometheus/prometheus",
        "severity": "warning"
      },
      "annotations": {
        "message": "30% throttling of CPU in namespace prometheus for container config-reloader in pod alertmanager-main-1.",
        "runbook_url": "https://github.com/kubernetes-monitoring/kubernetes-mixin/tree/master/runbook.md#alert-name-cputhrottlinghigh"
      },
      "startsAt": "2020-04-10T03:31:01.534463406Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://prometheus-prometheus-1:9090/graph?g0.expr=sum+by%28container%2C+pod%2C+namespace%29+%28increase%28container_cpu_cfs_throttled_periods_total%7Bcontainer%21%3D%22%22%7D%5B5m%5D%29%29+%2F+sum+by%28container%2C+pod%2C+namespace%29+%28increase%28container_cpu_cfs_periods_total%5B5m%5D%29%29+%3E+%2825+%2F+100%29&g0.tab=1",
      "fingerprint": "aebedebe331b5f7e"
    }
  ],
  "groupLabels": {
    "namespace": "prometheus"
  },
  "commonLabels": {
    "alertname": "CPUThrottlingHigh",
    "container": "config-reloader",
    "namespace": "prometheus",
    "prometheus": "prometheus/prometheus",
    "severity": "warning"
  },
  "commonAnnotations": {
    "runbook_url": "https://github.com/kubernetes-monitoring/kubernetes-mixin/tree/master/runbook.md#alert-name-cputhrottlinghigh"
  },
  "externalURL": "http://alertmanager-main-0:9093",
  "version": "4",
  "groupKey": "{}:{namespace=\"prometheus\"}"
}
`)

var Request = &http.Request{
	Host:   "localhost",
	Method: http.MethodPost,
	URL:    &url.URL{Host: "localhost:8080"},
	Proto:  "HTTP/2",
	Header: map[string][]string{
		"Accept-Encoding": {"gzip, deflate"},
		"Accept-Language": {"en-us"},
		"Foo":             {"Bar", "two"},
	},
	Body: ioutil.NopCloser(strings.NewReader(string(message))),
}

func TestRequests(t *testing.T) {
	t.Run("check if ping endpoint responds correctly", checkPing)
}

func TestMatrixBotName(t *testing.T) {
	displayName = "Omar" // Set flag variable

	var text = []byte(`junk`)
	var matrixMessage = &MatrixMessage{}
	j, err := encodeMessage(text)
	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(j, matrixMessage)
	if err != nil {
		t.Error(err)
	}

	want := displayName
	got := matrixMessage.DisplayName
	if got != want {
		t.Errorf("got %q wanted %q", got, want)
	}

}

func templateCheck(template []byte, message *HookMessage, t *testing.T) []byte {
	got, err := templateMessage(message, template)
	if err != nil {
		t.Error(err)
	}
	return got
}

func NewMockMessage() (*HookMessage, error) {
	message, err := decodeMessage(Request)
	if err != nil {
		return message, err
	}
	return message, nil
}

func TestTemplateMessage(t *testing.T) {
	var template []byte
	m, err := NewMockMessage()
	if err != nil {
		t.Errorf("error creating mock message: %v", err)
	}

	// with this template
	t.Logf("testing template which should print only the status")
	template = []byte(`{{ .Status -}}`)
	want := []byte(`firing`)
	got := templateCheck(template, m, t)
	if bytes.Compare(got, want) != 0 {
		t.Errorf("got %q wanted %q", got, want)
	}

	// with this template
	templateTwo := []byte(
		`{{ range .Alerts -}}
     {{ .Labels.alertname -}}
     {{ end -}}
     `)
	want = []byte(`CPUThrottlingHighCPUThrottlingHigh`)
	got = templateCheck(templateTwo, m, t)
	if bytes.Compare(got, want) != 0 {
		t.Errorf("got %q wanted %q", got, want)
	}

}

func TestDecodeMessage(t *testing.T) {
	var want interface{}
	var got interface{}

	m, err := NewMockMessage()
	if err != nil {
		t.Errorf("error calling decodeMessage: %v", err)
	}

	t.Log("test if first alert's severity has been decoded as `warning`")
	want = "warning"
	got = m.Alerts[0].Labels["severity"]
	if got != want {
		t.Errorf("got %q wanted %q", got, want)
	}

	want = 2
	got = len(m.Alerts)
	if got != want {
		t.Errorf("got %q wanted %q", got, want)
	}

	want = "firing"
	got = m.Status
	if got != want {
		t.Errorf("got %q wanted %q", got, want)
	}
}

func checkPing(t *testing.T) {
	resp, _ := http.Get("http://localhost:8080/ping")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.FailNow()
	}
	defer resp.Body.Close()

	want := "pong\n"
	got := string(body)

	if want != got {
		t.Errorf("got %q wanted %q", string(got), string(want))
		t.Fail()
	}
}
