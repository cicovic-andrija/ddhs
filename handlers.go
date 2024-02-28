package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type Page struct {
	SearchQuery string
	Runs        []*Run
	Run         *Run
}

func runsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("tmpl/runs.html", "tmpl/lead.html", "tmpl/trail.html")
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(w, &Page{Runs: runs}); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("tmpl/run.html", "tmpl/lead.html", "tmpl/trail.html")
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(w, &Page{Run: NewRun()}); err != nil {
		fmt.Printf("%v\n", err)
	}
}

type HandlerMux interface {
	Handle(pattern string, handler http.Handler)
}

func register(mux HandlerMux) HandlerMux {
	if mux == nil {
		return nil
	}

	mux.Handle(
		"/runs",
		http.HandlerFunc(runsHandler),
	)

	mux.Handle(
		"/runs/new",
		http.HandlerFunc(runHandler),
	)

	return mux
}
