package config

import (
	"flag"
	"os"
)

// Flags holds the config as passed in via flags - duh
var Flags Config

type Config struct {
	Port          int
	TemplatePath  string
	UpstreamHost  string
	AvatarURL     string
	MessageFormat string
	DisplayName   string
}

func (c *Config) Read() *Config {
	c.Parse()
	return c
}

func (c *Config) Parse() {
	flag.StringVar(&Flags.UpstreamHost, "upstream", "http://webhook.chatops.svc.cluster.local:9000", "the webhook which recieves the message")
	flag.IntVar(&Flags.Port, "port", 8080, "which port ot listen on")
	flag.StringVar(&Flags.TemplatePath, "template", "message.tmpl", "path to Go template")
	flag.StringVar(&Flags.AvatarURL, "avatar", "https://i.imgur.com/IDOBtEJ.png", "URL of avatar to use")
	flag.StringVar(&Flags.MessageFormat, "format", "html", "html or plain formatting of messages")
	flag.StringVar(&Flags.DisplayName, "name", "Alertmanager", "name of the bot")
	flag.Parse()
}

func (c *Config) Template() (*os.File, error) {
	f, err := os.Open(c.TemplatePath)
	if err != nil {
		return nil, err
	}
	return f, nil
}
