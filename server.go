package main

import (
	"fmt"
	"os"

	"github.com/cicovic-andrija/libgo/https"
	"github.com/cicovic-andrija/libgo/logging"
)

// Top-level API to the server itself.

var server *https.HTTPSServer

func shutdown() {
	if err := server.Shutdown(); err != nil {
		crash("failure during server shutdown: %v", err)
	}
	os.Exit(0)
}

func crash(format string, v ...any) {
	err := fmt.Errorf(format, v...)
	trace(logging.SevError, err.Error())
	panic(err)
}

func trace(sev logging.Severity, format string, v ...any) {
	server.GetLogger().Output(sev, 2, format, v...)
}
