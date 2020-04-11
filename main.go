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
	upstreamHost  string
	templatePath  string
	avatarURL     string
	messageFormat string
	displayName   string
	httpPort      int
	// TemplateBuffer holds the loaded template
	TemplateBuffer []byte
)

func init() {
	flag.StringVar(&upstreamHost, "upstream", "http://localhost:9000", "the webhook which recieves the message")
	flag.IntVar(&httpPort, "port", 8080, "which port ot listen on")
	flag.StringVar(&templatePath, "template", "message.tmpl", "path to Go template")
	flag.StringVar(&avatarURL, "avatar", "https://i.imgur.com/IDOBtEJ.png", "URL of avatar to use")
	flag.StringVar(&messageFormat, "format", "html", "html or plain formatting of messages")
	flag.StringVar(&displayName, "name", "Alertmanager", "name of the bot")
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
	path = "/api/v1/matrix/hook/"
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
	log.Printf("starting: %v", httpPort)
	http.HandleFunc(path, handler)
	http.HandleFunc("/ping", ping)

	http.ListenAndServe(":8080", nil)
}

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
		Format:      messageFormat,
		AvatarURL:   avatarURL,
		DisplayName: displayName,
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

func ping(w http.ResponseWriter, r *http.Request) {
	log.Printf("ping from %v (%v)", r.Host, r.UserAgent())
	http.Error(w, "pong", 200)
}

func handler(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimLeft(r.URL.Path[1:], path)

	if r.Method != http.MethodPost {
		log.Printf("%v doesn't support %v", r.URL.Path[1:], r.Method)
		http.Error(w, "invalid method", 401)
		return
	}

	log.Printf("incoming from %v (%v)", r.Host, r.UserAgent())

	message, err := decodeMessage(r)
	if err != nil {
		http.Error(w, "invalid request body", 401)
		log.Printf("couldn't decode incoming webhook: %v", err)
		return
	}

	templatedMessage, err := templateMessage(message)
	if err != nil {
		http.Error(w, "invalid request body", 401)
		log.Printf("couldn't template message: %v", err)
		return
	}

	jsonResponse, err := encodeMessage(templatedMessage)
	if err != nil {
		http.Error(w, "couldn't marshal JSON response", 401)
		log.Printf("couldn't encode JSON: %v", err)
		return
	}

	response, err := sendMessage(jsonResponse, token)
	if err != nil {
		http.Error(w, "upstream unavailable", 401)
	}

	log.Printf("successfully relayed message to %v", upstreamHost)
	http.Error(w, "success", 200)
	fmt.Print(string(jsonResponse))
	response.Body.Close()
}
