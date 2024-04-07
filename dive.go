package main

import (
	"fmt"
	"time"
)

// Dive is an in-memory model of a single dive.
type Dive struct {
	id            string
	ix            int
	Data          *DiveRecord
	DateTimeIn    time.Time `json:"-"`
	DateTimeInStr string    `json:"date_time_in"`
	Site          string    `json:"site"`
	Duration      Duration  `json:"duration"`
}

// NewDive returns a new, empty model, with all fields initialized to their "zero" or default/invalid value.
func NewDive() *Dive {
	return &Dive{ix: InvalidIndex}
}

// SetDateTimeAndAssignID must be used once to assign an ID (derived from the dive date and time) to a new model.
func (d *Dive) SetDateTimeAndAssignID(new time.Time) {
	if d.id == "" {
		d.DateTimeIn = new
		d.DateTimeInStr = d.DateTimeIn.Format(DateTimeLayout)
		d.id = d.DateTimeIn.Format(URLFriendlyDateTimeLayout)
	}
}

// ID returns the ID of the dive.
func (d *Dive) ID() string {
	return d.id
}

// Num returns the cardinal number of the dive, as set by its `DiveLog`.
func (d *Dive) Num() int {
	return d.ix + 1
}

// TimeOut returns the calculated date and time of the end of the dive.
func (d *Dive) TimeOut() time.Time {
	return d.DateTimeIn.Add(d.Duration.Value())
}

// Reconstruct the complete state based on partially initialized set of fields.
// TODO: Additional validation?
func (d *Dive) Reconstruct() error {
	dti, err := time.Parse(DateTimeLayout, d.DateTimeInStr)
	if err != nil {
		return fmt.Errorf("invalid date time in: %v", err)
	}
	d.SetDateTimeAndAssignID(dti)
	return nil
}
