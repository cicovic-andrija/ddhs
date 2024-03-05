package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cicovic-andrija/libgo/logging"
)

type Page struct {
	Title        string
	BeforeFilter string
	AfterFilter  string
	InputErrors  map[string]string
	Dive         *Dive
	Dives        []*Dive
}

func divesHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/") {
		redirect(strings.TrimSuffix(r.URL.Path, "/"), w, r)
		return
	}

	page := &Page{
		Title:        "Dive Log",
		BeforeFilter: r.URL.Query().Get("before"),
		AfterFilter:  r.URL.Query().Get("after"),
		Dives:        dives,
	}

	render("dives.html", w, page)
}

func diveHandler(w http.ResponseWriter, r *http.Request) {
	num, err := strconv.Atoi(r.PathValue("num"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	id := ntoi(num)
	if id < 0 || id >= len(dives) {
		http.NotFound(w, r)
		return
	}

	dive := dives[id]
	page := &Page{
		Title: fmt.Sprintf("Dive #%d", dive.Num),
		Dive:  dive,
	}

	render("dive.html", w, page)
}

func newDiveHandler(w http.ResponseWriter, r *http.Request) {
	page := &Page{
		Title: "New Dive",
		Dive:  NewDive(),
	}

	render("dive.html", w, page)
}

func diveFormHandler(w http.ResponseWriter, r *http.Request) {
	var (
		id int
		// dive = NewDive()
	)

	if numStr := r.PathValue("num"); numStr != "" { // .../{num}/edit
		if num, err := strconv.Atoi(r.PathValue("num")); err != nil {
			http.NotFound(w, r)
			return
		} else {
			id = ntoi(num)
			if id < 0 || id >= len(dives) {
				http.NotFound(w, r)
				return
			}
		}
	} else { // .../new
		id = -1
	}

	redirect("/dives", w, r)
}

func render(tmplName string, w http.ResponseWriter, data any) {
	const (
		tmplDir = "tmpl"
	)
	tmpl, err := template.ParseFiles(filepath.Join(tmplDir, tmplName), filepath.Join(tmplDir, "partials.html"))
	if err != nil {
		traceServerMessage(logging.SevError, "%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		traceServerMessage(logging.SevError, "%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func parseDateFromRequest(r *http.Request, nameInForm string) (date time.Time, err error) {
	if dateStr := r.FormValue(nameInForm); dateStr != "" {
		if parsed, pe := time.Parse("2006-01-02", dateStr); pe != nil {
			err = errors.New("provided date isn't valid")
		} else {
			date = parsed
		}
	} else {
		err = errors.New("date not provided")
	}
	return
}

type HandlerMux interface {
	Handle(pattern string, handler http.Handler)
}

func register(mux HandlerMux) HandlerMux {
	if mux == nil {
		return nil
	}

	mux.Handle(
		"GET /dives",
		http.HandlerFunc(divesHandler),
	)

	mux.Handle(
		"GET /dives/{$}",
		http.HandlerFunc(divesHandler),
	)

	mux.Handle(
		"GET /dives/{num}",
		http.HandlerFunc(diveHandler),
	)

	mux.Handle(
		"POST /dives/{num}/edit",
		http.HandlerFunc(diveFormHandler),
	)

	mux.Handle(
		"GET /dives/new",
		http.HandlerFunc(newDiveHandler),
	)

	mux.Handle(
		"POST /dives/new",
		http.HandlerFunc(diveFormHandler),
	)

	return mux
}

func redirect(path string, w http.ResponseWriter, r *http.Request) {
	u := &url.URL{Path: path, RawQuery: r.URL.RawQuery}
	http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
}
