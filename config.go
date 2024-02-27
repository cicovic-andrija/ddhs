package main

import "flag"

var config struct {
	host string
	port int
}

func parseArgs() {
	var (
		devFlag = flag.Bool("d", false, "dev (local) run")
	)

	flag.Parse()
	config.host = "any"
	config.port = 443
	if *devFlag {
		config.host = "localhost"
		config.port = 8080
	}
}
