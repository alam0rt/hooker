package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/alam0rt/hooker/config"
	"github.com/alam0rt/hooker/message"
)

const (
	path = "/api/v1/matrix/hook/"
)

// Start initialises the HTTP server
func Start() {
	log.Printf("starting server on port %d", config.Flags.Port)
	http.HandleFunc("/ping", ping)
	http.HandleFunc(path, ingest)
	http.ListenAndServe(":"+string(config.Flags.Port), nil)
}

func ingest(w http.ResponseWriter, r *http.Request) {
	var m message.Message

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	err = m.Webhook.Unmarshal(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	//TODO: make this less shit.
	m.Template.File, err = os.Open("../message.tmpl")
	fmt.Print(err)
	err = m.Render()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	reply, err := m.Matrix.Marshal()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(reply)
	}

	w.WriteHeader(http.StatusOK)
}

func ping(w http.ResponseWriter, r *http.Request) {
	response := []byte(`pong`)
	w.WriteHeader(http.StatusOK)
	fmt.Print(r)
	w.Write(response)
}
