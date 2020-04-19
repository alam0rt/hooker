package message

import (
	"encoding/json"
	"os"
	"sync"
	tmpl "text/template"
	"time"
)

type (

	// Timestamp is a helper for (un)marhalling time
	Timestamp time.Time

	// Message is a generic type encompassing both webook messages & messages destined for Matrix
	Message struct {
		Webhook  HookMessage
		Matrix   MatrixMessage
		Template Template
	}

	// Template is just that
	Template struct {
		body []byte
		File *os.File
	}

	// HookMessage is the message we receive from Alertmanager
	HookMessage struct {
		Version           string            `json:"version"`
		GroupKey          string            `json:"groupKey"`
		Status            string            `json:"status"`
		Receiver          string            `json:"receiver"`
		GroupLabels       map[string]string `json:"groupLabels"`
		CommonLabels      map[string]string `json:"commonLabels"`
		CommonAnnotations map[string]string `json:"commonAnnotations"`
		ExternalURL       string            `json:"externalURL"`
		Alerts            []Alert           `json:"alerts"`
	}

	// MatrixMessage is the message format we send to Matrix
	MatrixMessage struct {
		Text        string `json:"text"`
		Format      string `json:"format"`
		DisplayName string `json:"displayName"`
		AvatarURL   string `json:"avatarUrl"`
	}

	// Alert is a single alert.
	Alert struct {
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		StartsAt    string            `json:"startsAt,omitempty"`
		EndsAt      string            `json:"EndsAt,omitempty"`
	}

	// just an example alert store. in a real hook, you would do something useful
	alertStore struct {
		sync.Mutex
		capacity int
		alerts   []*HookMessage
	}
)

// Marshal returns the JSON formatted message
func (m *MatrixMessage) Marshal() ([]byte, error) {
	j, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// Write implements the io.Writer interface
func (m *MatrixMessage) Write(data []byte) (n int, err error) {
	n = len(data)
	m.Text = string(data)
	return n, nil
}

// Render takes the webhook and formats it into a Matrix Message
func (m *Message) Render() error {
	err := m.Template.readTemplate()
	if err != nil {
		return err
	}
	template, err := tmpl.New("message").Parse(string(m.Template.body))
	if err != nil {
		return err
	}
	template.Execute(&m.Matrix, m.Webhook)
	if err != nil {
		return err
	}
	return nil
}

// Unmarshal takes a webhook byte array and populates the struct
func (m *HookMessage) Unmarshal(b []byte) error {
	err := json.Unmarshal(b, m)
	if err != nil {
		return err
	}
	return nil
}

func (t *Template) readTemplate() error {
	data := make([]byte, 1024)
	_, err := t.File.Read(data)
	if err != nil {
		return err
	}

	return nil
}
