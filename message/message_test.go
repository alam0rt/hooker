package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func newMockWebhook() *HookMessage {
	wh := &HookMessage{}
	wh.Unmarshal(MockHook)
	return wh
}

func newMockMessage() *Message {
	wh := newMockWebhook()
	msg := &MatrixMessage{
		DisplayName: "mock",
		AvatarURL:   "http://mock.png",
		Text:        "",
	}
	m := &Message{
		Webhook: *wh,
		Matrix:  *msg,
	}
	return m
}

func TestUnmarshal(t *testing.T) {
	wh := &HookMessage{}
	wh.Unmarshal(MockHook)
	assert.Equal(t, "firing", wh.Status, "the status should be firing")

	assert.Equal(t, 2, len(wh.Alerts), "there should be 2 alerts")

}

func TestRender(t *testing.T) {
	m := newMockMessage()
	m.Template.body = []byte(`{{ .Status -}}`)
	m.Render()
	t.Log(m.Matrix)
	assert.Equal(t, "firing", m.Matrix.Text)
}
