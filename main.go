package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/cicovic-andrija/libgo/fs"
	"github.com/cicovic-andrija/libgo/https"
	"github.com/cicovic-andrija/libgo/logging"
)

func main() {
	parseArgs()
	server := newHTTPSS()
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

func newHTTPSS() *https.HTTPSServer {
	if err := fs.MkdirIfNotExists("logs"); err != nil {
		crash("mkdir: %v", err)
	}

	server, err := https.NewServer(&https.Config{
		Network: https.NetworkConfig{
			IPAcceptHost: config.host,
			TCPPort:      config.port,
			TLSCertPath:  "local_assets/tlspublic.crt",
			TLSKeyPath:   "local_assets/tlsprivate.key",
		},
		EnableFileServer: true,
		FileServer: https.FileServerConfig{
			URLPrefix: "/static/",
			Directory: "static",
			Allowed:   []string{"site.css"},
		},
		LogRequests:   true,
		LogsDirectory: "logs",
	})
	if err != nil {
		crash("https: %v", err)
	}

	return server
}

func crash(format string, v ...any) {
	if crashlog, err := logging.NewFileLog("crash.log"); err == nil {
		crashlog.Output(logging.SevError, 2, format, v...)
	}
	panic(fmt.Errorf(format, v...))
}
