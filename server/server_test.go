package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const mockURL = "http://localhost:8080"

var MockHook = []byte(`
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

func TestPing(t *testing.T) {
	var expected interface{}
	var got interface{}

	r, err := http.NewRequest("GET", mockURL+"/ping", nil)
	assert.Nil(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)

	handler.ServeHTTP(rr, r)

	expected = http.StatusOK
	got = rr.Code
	assert.Equal(t, expected, got)

	expected = []byte(`pong`)
	got = rr.Body.Bytes()
	assert.Equal(t, expected, got)

}

func TestPostWebhook(t *testing.T) {
	var expected interface{}
	var got interface{}
	var token = "test"
	var body bytes.Buffer

	// populate body with mock webhook
	body.Write(MockHook)

	r, err := http.NewRequest("POST", mockURL+path+token, &body)
	assert.Nil(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ingest)

	handler.ServeHTTP(rr, r)

	expected = http.StatusOK
	got = rr.Code
	assert.Equal(t, expected, got)

	assert.NotNil(t, rr.Body, "the reply message body is nil")
	t.Log(rr.Body)

}
