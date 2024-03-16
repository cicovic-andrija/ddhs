package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

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

// Error messages returned by validation functions defined below are client-facing.
//

func validateDateInput(dateStr string) (date time.Time, errMsg string) {
	date, err := time.Parse(DateLayout, dateStr)
	if err != nil {
		errMsg = "Please provide a valid dive date."
	}
	return
}

func validateTimeInput(timeStr string) (t time.Time, errMsg string) {
	t, err := time.Parse(TimeLayout, timeStr)
	if err != nil {
		errMsg = "Please provide a valid dive start time."
	}
	return
}

func validateDiveSiteInput(inputStr string) (site string, errMsg string) {
	if site = strings.TrimSpace(inputStr); site == "" {
		errMsg = "Please name a location of the dive."
	}
	return
}

func validateDiveDurationInput(durationStr string) (d time.Duration, errMsg string) {
	return
}
