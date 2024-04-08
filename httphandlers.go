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
	SyncJob      *SyncJob
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
		Dive:  EmptyDive(),
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
		// TODO: Date and time must match between "new" and existing dive.
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
			if dive == nil {
				dive = EmptyDive()
			}
			page.Dive = dive
		}
		render("dive.html", w, page)
	}
}

func parseDiveFromRequest(r *http.Request, errorMap map[string]string) (dive *Dive, ok bool) {
	var (
		dt         time.Time
		diveRecord *DiveRecord
	)

	ok = true

	// Date and time in are parsed first, so a dive object can be initialized.
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

	dive = NewDive(dt) // this is fine even if ok == false at this point, collect other errors if any
	diveRecord = dive.Data

	if site, errMsg := validateDiveSiteInput(r.FormValue(SiteTag)); errMsg != "" {
		ok = false
		errorMap[SiteTag] = errMsg
	} else {
		diveRecord.Site = site
	}

	if d, errMsg := validateDurationInMinInput(r.FormValue(DurationTag)); errMsg != "" {
		ok = false
		errorMap[DurationTag] = errMsg
	} else {
		diveRecord.Duration = Duration{Duration: d}
	}

	// Optional parameter: accept even an empty value after input is trimmed.
	diveRecord.Geo, _ = validateNonEmptyString(r.FormValue(GeoTag))
	diveRecord.DecoDive = r.FormValue(DecoDiveTag) == "true"

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
	case DurationTag:
		_, errMsg = validateDurationInMinInput(value)
	case GeoTag:
		if value != "" { // optional, so empty value is not an error
			_, errMsg = validateNonEmptyString(value)
		}
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

func render(tmplName string, w http.ResponseWriter, data any) {
	tmpl, err := template.ParseFiles(filepath.Join(TmplDir, tmplName), filepath.Join(TmplDir, "partials.html"))
	if err != nil {
		trace(logging.SevError, "%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		trace(logging.SevError, "%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func partialRender(tmplName string, w http.ResponseWriter, data any) {
	partials, err := template.ParseFiles(filepath.Join(TmplDir, "partials.html"))
	if err != nil {
		trace(logging.SevError, "%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := partials.ExecuteTemplate(w, tmplName, data); err != nil {
		trace(logging.SevError, "%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func dateToStr(d time.Time) string {
	return d.Format(DateLayout)
}
