package main

import (
	"github.com/alam0rt/hooker/config"
	"github.com/alam0rt/hooker/server"
)

func main() {
	config.Flags.Parse()
	server.Start()
}
