package main

import (
	"fmt"
	"net/http"
)

func divesHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, world")
}

type HandlerMux interface {
	Handle(pattern string, handler http.Handler)
}

func register(mux HandlerMux) HandlerMux {
	if mux == nil {
		return nil
	}

	mux.Handle(
		"/dives",
		http.HandlerFunc(divesHandler),
	)

	return mux
}
