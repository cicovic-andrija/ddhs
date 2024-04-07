package main

import (
	"strings"
	"time"
)

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
