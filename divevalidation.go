package main

import (
	"strconv"
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
		errMsg = "Please provide the name of the dive site."
	}
	return
}

func validateDurationInMinInput(inputStr string) (d time.Duration, errMsg string) {
	mins, err := strconv.Atoi(inputStr)
	if err != nil {
		errMsg = "Please provide a valid duration in minutes."
	} else if mins < 1 || mins > 180 {
		errMsg = "Dive duration must be between 1 and 180 minutes."
	} else {
		d = time.Duration(mins) * time.Minute
	}
	return
}
