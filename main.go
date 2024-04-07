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
)

func main() {
	// Initialize.
	parseArgv()
	server = newHTTPSS()
	register(server)
	go ensureDataLoadAsync()

	// Ctrl-C handler.
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)

	// Start listening.
	errors := make(chan error, 1)
	server.ListenAndServeAsync(errors)

	// Halt the main thread until interrupt or unexpected failure.
	select {
	case <-interrupts:
		shutdown()
	case err := <-errors:
		crash("critical server failure: %v", err)
	}
}

func parseArgv() {
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
	const LogsDirectory = "logs"
	for _, dir := range []string{LogsDirectory, DataDirectory} {
		if err := fs.MkdirIfNotExists(dir); err != nil {
			crashEarly("mkdir: %v", err)
		}
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
			Allowed:   []string{"site.css", "simple2.3.0.min.css"},
		},
		LogRequests:   config.logRequests,
		LogsDirectory: LogsDirectory,
	})
	if err != nil {
		crashEarly("https: %v", err)
	}

	return srv
}

func crashEarly(format string, v ...any) {
	if crashFile, err := logging.NewFileLog("crash.log"); err == nil {
		crashFile.Output(logging.SevError, 2, format, v...)
	}
	panic(fmt.Errorf(format, v...))
}
