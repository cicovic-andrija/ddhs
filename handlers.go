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

const (
	PageSize = 10

	PageQueryTag   = "page"
	BeforeQueryTag = "before"
	AfterQueryTag  = "after"
)

type Page struct {
	Title        string
	BeforeFilter time.Time
	AfterFilter  time.Time
	PageFilter   int
	LastPage     bool
	InputErrors  map[string]string
	Dive         *Dive
	Dives        []*Dive
	Total        int
	Renumbered   bool
}

func (p *Page) NextPage() int {
	return p.PageFilter + 1
}

func (p *Page) BeforeQueryVal() string {
	if p.BeforeFilter.IsZero() {
		return ""
	}
	return dateToStr(p.BeforeFilter)
}

func (p *Page) URLBeforeQuery() string {
	if p.BeforeFilter.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s=%s&", BeforeQueryTag, dateToStr(p.BeforeFilter))
}

func (p *Page) AfterQueryVal() string {
	if p.AfterFilter.IsZero() {
		return ""
	}
	return dateToStr(p.AfterFilter)
}

func (p *Page) URLAfterQuery() string {
	if p.AfterFilter.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s=%s&", AfterQueryTag, dateToStr(p.AfterFilter))
}

func divesHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/") {
		u := &url.URL{Path: strings.TrimSuffix(r.URL.Path, "/"), RawQuery: r.URL.RawQuery}
		http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
		return
	}

	page := &Page{Title: "Dive Log"}

	InMemLog.RLock()
	defer InMemLog.RUnlock()
	filtered := InMemLog.All()

	if beforeValue := r.URL.Query().Get(BeforeQueryTag); beforeValue != "" {
		if beforeDate, err := time.Parse(DateLayout, beforeValue); err == nil {
			page.BeforeFilter = beforeDate
			filtered = filtered.Filter(func(dive *Dive) bool { return dive.Date.Before(beforeDate) })
		}
	}

	if afterValue := r.URL.Query().Get(AfterQueryTag); afterValue != "" {
		if afterDate, err := time.Parse(DateLayout, afterValue); err == nil {
			page.AfterFilter = afterDate
			filtered = filtered.Filter(func(dive *Dive) bool { return dive.Date.After(afterDate) })
		}
	}

	if InMemLog.IsRenumbered() {
		page.Renumbered = true
	}

	page.Total = len(filtered)
	pageNum, err := strconv.Atoi(r.URL.Query().Get(PageQueryTag))
	if err != nil || pageNum < 1 {
		pageNum = 1
	}
	page.PageFilter = pageNum
	filtered = Paginate(filtered, pageNum-1, PageSize)

	page.Dives = filtered
	page.LastPage = len(filtered) < PageSize // there is an acceptable fencepost error here
	render("dives.html", w, page)
}

func diveHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue(IDTag)
	InMemLog.RLock()
	defer InMemLog.RUnlock()

	dive := InMemLog.Find(id)
	if dive == nil {
		http.NotFound(w, r)
		return
	}

	page := &Page{
		Title: fmt.Sprintf("Dive #%d", dive.Num()),
		Dive:  dive,
	}

	render("dive.html", w, page)
}

func diveRemovalHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue(IDTag)
	InMemLog.Lock()
	defer InMemLog.Unlock()
	InMemLog.Delete(id)

	// TODO: Also check for HX-Request header.
	if r.Header.Get("HX-Trigger") == "delete-btn" {
		http.Redirect(w, r, "/dives", http.StatusSeeOther) // 303 because this is a redirect to a DELETE request
	} else {
		fmt.Fprint(w, "")
	}
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
		page     = &Page{InputErrors: make(map[string]string)}
		existing *Dive
	)

	if id := r.PathValue(IDTag); id != "" { // .../{id}/edit
		InMemLog.RLock()
		if existing = InMemLog.Find(id); existing == nil {
			InMemLog.RUnlock()
			http.NotFound(w, r)
			return
		}
		InMemLog.RUnlock()
	} // else .../new

	if dive, ok := parseDiveFromRequest(r, page.InputErrors); ok {
		InMemLog.Lock()
		if existing != nil {
			InMemLog.Replace(existing, dive)
		} else {
			InMemLog.Insert(dive)
		}
		InMemLog.Unlock()
		http.Redirect(w, r, "/dives", http.StatusFound)
	} else { // not ok
		page.Title = "New Dive"
		if existing != nil {
			page.Title = fmt.Sprintf("Dive #%d", existing.Num())
			page.Dive = existing
		} else {
			page.Dive = dive
		}
		render("dive.html", w, page)
	}
}

func parseDiveFromRequest(r *http.Request, errorMap map[string]string) (dive *Dive, ok bool) {
	ok = true
	dive = NewDive()
	dive.Site = r.FormValue(SiteTag)

	if date, err := parseDateFromRequest(r, DateTag); err != nil {
		ok = false
		errorMap[DateTag] = fmt.Sprintf("Invalid input: %v.", err)
	} else {
		dive.Date = date
	}

	return
}

func inputValidationHandler(w http.ResponseWriter, r *http.Request) {
	var (
		errMsg = ""
		tag    = r.PathValue("tag")
		value  = r.URL.Query().Get(tag)
	)

	switch tag {
	case SiteTag:
		if strings.TrimSpace(value) == "" {
			errMsg = "Please name a <mark>location</mark> of the dive."
		}
	}

	fmt.Fprintf(w, "%s", errMsg)
}

func parseDateFromRequest(r *http.Request, tag string) (date time.Time, err error) {
	if dateStr := r.FormValue(tag); dateStr != "" {
		if parsed, pe := time.Parse(DateLayout, dateStr); pe != nil {
			err = errors.New("provided date isn't valid")
		} else {
			date = parsed
		}
	} else {
		err = errors.New("date not provided")
	}
	return
}

func dateToStr(d time.Time) string {
	return d.Format(DateLayout)
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

	return mux
}
