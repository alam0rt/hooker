package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"text/template"
	"time"
)

var (
	upstreamHost   string
	templatePath   string
	httpPort       int
	TemplateBuffer []byte
)

func init() {
	flag.StringVar(&upstreamHost, "upstream", "http://localhost:3070", "the webhook which recieves the message")
	flag.IntVar(&httpPort, "port", 8080, "which port ot listen on")
	flag.StringVar(&templatePath, "template", "message.tmpl", "path to Go template")
}

func loadTemplate() error {

	var err error
	TemplateBuffer, err = ioutil.ReadFile(templatePath)
	if err != nil {
		return err
	}
	return nil
}

const (
	path      = "/api/v1/matrix/hook/"
	avatarURL = "https://i.imgur.com/IDOBtEJ.png"
)

type (

	// Timestamp is a helper for (un)marhalling time
	Timestamp time.Time

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

func main() {
	flag.Parse()
	err := loadTemplate()
	if err != nil {
		fmt.Printf("unable to open template: %v", err)
	}
	fmt.Print(TemplateBuffer)
	log.Printf("starting: %v", httpPort)
	http.HandleFunc(path, handler)

	http.ListenAndServe(":8080", nil)
}

var tmpl = []byte(`
[{{ .Status }}]
{{ range .Alerts }}
  {{ .StartsAt }}
{{ end }}
`)

func decodeMessage(r *http.Request) (m *HookMessage, e error) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	m = &HookMessage{}
	if err := dec.Decode(m); err != nil {
		log.Printf("error decoding message: %v", err)
		return nil, err
	}
	return m, nil

}

func templateMessage(m *HookMessage) (b []byte, err error) {
	var buf bytes.Buffer
	tmpl, err := template.New("template").Parse(string(TemplateBuffer))
	if err != nil {
		log.Printf("couldn't template message: %v", err)
		return nil, err
	}

	err = tmpl.Execute(&buf, m)
	if err != nil {
		log.Printf("couldn't parse template: %v", err)
		return nil, err

	}
	return buf.Bytes(), err
}

func encodeMessage(b []byte) ([]byte, error) {
	m := &MatrixMessage{
		Text:        string(b),
		Format:      "test",
		AvatarURL:   avatarURL,
		DisplayName: "bot",
	}

	j, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("error converting response to JSON: %v", err)
		return nil, err
	}
	return j, nil
}

func sendMessage(m []byte, token string) (resp *http.Response, e error) {
	host := upstreamHost + path + token
	resp, err := http.Post(host, "application/json", bytes.NewBuffer(m))
	if err != nil {
		fmt.Printf("couldn't send message to upstream: %v", err)
	}
	defer resp.Body.Close()

	return resp, nil

}

func handler(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimLeft(r.URL.Path[1:], path)
	fmt.Fprintf(w, "%s", r.URL.Path[1:])

	message, err := decodeMessage(r)
	if err != nil {
		http.Error(w, "invalid request body", 401)
		log.Printf("couldn't decode incoming webhook: %v", err)
	}

	templatedMessage, err := templateMessage(message)
	if err != nil {
		http.Error(w, "invalid request body", 401)
	}

	jsonResponse, err := encodeMessage(templatedMessage)
	if err != nil {
		http.Error(w, "couldn't marshal JSON response", 401)
	}

	response, err := sendMessage(jsonResponse, token)
	if err != nil {
		http.Error(w, "upstream unavailable", 401)
	}
	response.Body.Close()
}
