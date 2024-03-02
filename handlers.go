package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type Page struct {
	SearchQuery string

	Dive  ModelRepresentation[Dive]
	Dives []ModelRepresentation[Dive]

	Run  ModelRepresentation[Run]
	Runs []ModelRepresentation[Run]
}

type ModelRepresentation[T any] struct {
	Data        *T
	InputErrors map[string]string
}

func WrapModel[T any](ptr *T) ModelRepresentation[T] {
	return ModelRepresentation[T]{
		Data:        ptr,
		InputErrors: make(map[string]string),
	}
}

func WrapMultipleModels[T any](s []*T) []ModelRepresentation[T] {
	fmt.Println(len(s))
	wrapped := make([]ModelRepresentation[T], 0, len(s))
	for _, ptr := range s {
		wrapped = append(wrapped, WrapModel(ptr))
	}
	return wrapped
}

func runsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("tmpl/runs.html", "tmpl/lead.html", "tmpl/trail.html")
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(w, &Page{Runs: WrapMultipleModels(runs)}); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("tmpl/run.html", "tmpl/lead.html", "tmpl/trail.html")
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(w, &Page{Run: WrapModel(NewRun())}); err != nil {
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
