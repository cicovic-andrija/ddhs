package main

import (
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
	TmplDir = "tmpl"

	PageSize = 10

	PageQueryTag   = "page"
	BeforeQueryTag = "before"
	AfterQueryTag  = "after"
)

// TODO: Should be used only when rendering a whole page template
type Page struct {
	Title           string
	BeforeFilter    time.Time
	AfterFilter     time.Time
	PageFilter      int
	LastPage        bool
	InputErrors     map[string]string
	Dive            *Dive
	Dives           []*Dive
	Total           int
	Renumbered      bool
	PersistenceInfo string
	SyncJob         *SyncJob
}

func (p *Page) NextPage() int {
	return p.PageFilter + 1
}

func (p *Page) URLBeforeQuery() string {
	if p.BeforeFilter.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s=%s&", BeforeQueryTag, dateToStr(p.BeforeFilter))
}

func (p *Page) URLAfterQuery() string {
	if p.AfterFilter.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s=%s&", AfterQueryTag, dateToStr(p.AfterFilter))
}

func (p *Page) NormalizedDateValue(date time.Time) string {
	if date.IsZero() {
		return ""
	}
	return dateToStr(date)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	u := &url.URL{Path: "/dives", RawQuery: r.URL.RawQuery}
	http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
}

func divesHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/") {
		u := &url.URL{Path: strings.TrimSuffix(r.URL.Path, "/"), RawQuery: r.URL.RawQuery}
		http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
		return
	}

	page := &Page{Title: "Dive Log"}

	MLog.RLock()
	defer MLog.RUnlock()
	filtered := MLog.All()

	if beforeValue := r.URL.Query().Get(BeforeQueryTag); beforeValue != "" {
		if beforeDate, err := time.Parse(DateLayout, beforeValue); err == nil {
			page.BeforeFilter = beforeDate
			filtered = filtered.Filter(func(dive *Dive) bool { return dive.DateTimeIn.Before(beforeDate) })
		}
	}

	if afterValue := r.URL.Query().Get(AfterQueryTag); afterValue != "" {
		if afterDate, err := time.Parse(DateLayout, afterValue); err == nil {
			page.AfterFilter = afterDate
			filtered = filtered.Filter(func(dive *Dive) bool { return dive.DateTimeIn.After(afterDate) })
		}
	}

	if MLog.IsRenumbered() {
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
	page.PersistenceInfo = fmt.Sprintf("seq::%d::%s", MLog.sequence, MLog.lastPersisted.Format(time.RFC3339))
	page.SyncJob = syncJob
	render("dives.html", w, page)
}

func diveHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue(IDTag)
	MLog.RLock()
	defer MLog.RUnlock()

	dive := MLog.Find(id)
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
	MLog.Lock()
	defer MLog.Unlock()
	MLog.Delete(id)

	// TODO: Also check for HX-Request header.
	if r.Header.Get("HX-Trigger") == "delete-btn" {
		http.Redirect(w, r, "/dives", http.StatusSeeOther) // 303 because this is a redirect to a DELETE request
	} else {
		fmt.Fprint(w, "") // this is an async call, return hypermedia in the response to client
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
		MLog.RLock()
		if existing = MLog.Find(id); existing == nil {
			MLog.RUnlock()
			http.NotFound(w, r)
			return
		}
		MLog.RUnlock()
	} // else .../new

	if dive, ok := parseDiveFromRequest(r, page.InputErrors); ok {
		// TODO: date and time must match between "new" and existing dive
		MLog.Lock()
		if existing != nil {
			MLog.Replace(existing, dive)
		} else {
			MLog.Insert(dive)
		}
		MLog.Unlock()
		http.Redirect(w, r, "/dives", http.StatusFound)
	} else { // not ok
		page.Title = "New Dive"
		if existing != nil {
			page.Title = fmt.Sprintf("Dive #%d", existing.Num())
			page.Dive = existing
		} else {
			page.Dive = dive
		}
		// TODO: Should response code be changed here?
		render("dive.html", w, page)
	}
}

func parseDiveFromRequest(r *http.Request, errorMap map[string]string) (dive *Dive, ok bool) {
	ok = true
	dive = NewDive()
	dt := dive.DateTimeIn // dt must be initialized to "zero" date/time, so copy from a new "zero" dive

	// Date and time in are parsed first, so the ID of the dive can be generated.
	if date, errMsg := validateDateInput(r.FormValue(DateTag)); errMsg != "" {
		ok = false
		errorMap[DateTag] = errMsg
	} else {
		dt = date
	}
	if timeIn, errMsg := validateTimeInput(r.FormValue(TimeInTag)); errMsg != "" {
		ok = false
		errorMap[TimeInTag] = errMsg
	} else {
		dt = dt.Add(time.Duration(timeIn.Hour())*time.Hour + time.Duration(timeIn.Minute())*time.Minute)
	}
	dive.SetDateTimeAndAssignID(dt)

	if site, errMsg := validateDiveSiteInput(r.FormValue(SiteTag)); errMsg != "" {
		ok = false
		errorMap[SiteTag] = errMsg
	} else {
		dive.Site = site
	}

	return
}

// HTTPS handler responsible for async. input validation. Returns hypermedia in the response to the client.
func inputValidationHandler(w http.ResponseWriter, r *http.Request) {
	var (
		tag    = r.PathValue("tag")
		value  = r.URL.Query().Get(tag)
		errMsg = ""
	)

	switch tag {
	case SiteTag:
		_, errMsg = validateDiveSiteInput(value)
	case DateTag:
		_, errMsg = validateDateInput(value)
	case TimeInTag:
		_, errMsg = validateTimeInput(value)
	}

	fmt.Fprintf(w, "%s", errMsg)
}

func syncHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		syncJob.Start()
		partialRender("sync-ui", w, syncJob)
	} else if r.Method == http.MethodGet {
		partialRender("sync-ui", w, syncJob)
	}
}

func dateToStr(d time.Time) string {
	return d.Format(DateLayout)
}

func render(tmplName string, w http.ResponseWriter, data any) {
	tmpl, err := template.ParseFiles(filepath.Join(TmplDir, tmplName), filepath.Join(TmplDir, "partials.html"))
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

func partialRender(tmplName string, w http.ResponseWriter, data any) {
	partials, err := template.ParseFiles(filepath.Join(TmplDir, "partials.html"))
	if err != nil {
		traceServerMessage(logging.SevError, "%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := partials.ExecuteTemplate(w, tmplName, data); err != nil {
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
