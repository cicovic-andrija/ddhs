package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/cicovic-andrija/libgo/fs"
	"github.com/cicovic-andrija/libgo/https"
	"github.com/cicovic-andrija/libgo/logging"
)

var (
	config struct {
		host        string
		port        int
		logRequests bool
	}

	server *https.HTTPSServer
)

func main() {
	parseArgs()
	server = newHTTPSS()
	register(server)
	errors := make(chan error, 1)
	interrupts := make(chan os.Signal, 1)
	server.ListenAndServeAsync(errors)
	signal.Notify(interrupts, os.Interrupt)
	for {
		select {
		case <-interrupts:
			if err := server.Shutdown(); err != nil {
				crash("failure during server shutdown: %v", err)
			}
			os.Exit(0)
		case err := <-errors:
			crash("critical server failure: %v", err)
		}
	}
}

func parseArgs() {
	var (
		devFlag = flag.Bool("d", false, "dev (local) execution")
	)

	flag.Parse()
	config.host = "any"
	config.port = 443
	config.logRequests = false
	if *devFlag {
		config.host = "localhost"
		config.port = 8080
		config.logRequests = true
	}
}

func newHTTPSS() *https.HTTPSServer {
	if err := fs.MkdirIfNotExists("logs"); err != nil {
		crashEarly("mkdir: %v", err)
	}
	if err := fs.MkdirIfNotExists(DataDirectory); err != nil {
		crashEarly("mkdir: %v", err)
	}

	srv, err := https.NewServer(&https.Config{
		Network: https.NetworkConfig{
			IPAcceptHost: config.host,
			TCPPort:      config.port,
			TLSCertPath:  "local_assets/tlspublic.crt",
			TLSKeyPath:   "local_assets/tlsprivate.key",
		},
		FileServer: https.FileServerConfig{
			Enabled:   true,
			URLPrefix: "/static/",
			Directory: "static",
			Allowed:   []string{"site.css"},
		},
		LogRequests:   config.logRequests,
		LogsDirectory: "logs",
	})
	if err != nil {
		crashEarly("https: %v", err)
	}

	return srv
}

func traceServerMessage(sev logging.Severity, format string, v ...any) {
	server.GetLogger().Output(sev, 2, format, v...)
}

func crash(format string, v ...any) {
	err := fmt.Errorf(format, v...)
	traceServerMessage(logging.SevError, err.Error())
	panic(err)
}

func crashEarly(format string, v ...any) {
	if crashFile, err := logging.NewFileLog("crash.log"); err == nil {
		crashFile.Output(logging.SevError, 2, format, v...)
	}
	panic(fmt.Errorf(format, v...))
}
