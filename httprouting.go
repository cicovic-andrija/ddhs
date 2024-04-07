package main

import "net/http"

type HandlerMux interface {
	Handle(pattern string, handler http.Handler)
}

func register(mux HandlerMux) HandlerMux {
	if mux == nil {
		return nil
	}

	mux.Handle(
		"GET /{$}",
		http.HandlerFunc(rootHandler),
	)

	mux.Handle(
		"GET /dives",
		http.HandlerFunc(divesHandler),
	)

	mux.Handle(
		"GET /dives/{$}",
		http.HandlerFunc(divesHandler),
	)

	mux.Handle(
		"GET /dives/{id}",
		http.HandlerFunc(diveHandler),
	)

	mux.Handle(
		"POST /dives/{id}/edit",
		http.HandlerFunc(diveFormHandler),
	)

	mux.Handle(
		"DELETE /dives/{id}",
		http.HandlerFunc(diveRemovalHandler),
	)

	mux.Handle(
		"GET /dives/new",
		http.HandlerFunc(newDiveHandler),
	)

	mux.Handle(
		"POST /dives/new",
		http.HandlerFunc(diveFormHandler),
	)

	mux.Handle(
		"GET /actions/validate/{tag}",
		http.HandlerFunc(inputValidationHandler),
	)

	mux.Handle(
		"POST /actions/sync",
		http.HandlerFunc(syncHandler),
	)

	mux.Handle(
		"GET /actions/sync",
		http.HandlerFunc(syncHandler),
	)

	return mux
}
