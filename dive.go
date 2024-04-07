package main

import (
	"fmt"
	"time"
)

// Dive is an in-memory model of a single dive.
type Dive struct {
	id         string
	ix         int
	DateTimeIn time.Time
	Data       *DiveRecord
}

// NewDive returns a new, initialized model, with non-ID fields initialized to their "zero"/default/invalid value.
func NewDive(dt time.Time) *Dive {
	return &Dive{
		id:         dt.Format(URLFriendlyDateTimeLayout),
		ix:         InvalidIndex,
		DateTimeIn: dt,
		Data: &DiveRecord{
			DateTime: dt.Format(DateTimeLayout),
		},
	}
}

// EmptyDive returns a new, empty model, with all fields initialized to their "zero"/default/invalid value.
func EmptyDive() *Dive {
	return &Dive{
		ix:   InvalidIndex,
		Data: &DiveRecord{},
	}
}

// ReconstructFrom initializes and empty
func (d *Dive) reconstructFrom(diveRecord *DiveRecord) (*Dive, error) {
	dt, err := time.Parse(DateTimeLayout, diveRecord.DateTime)
	if err != nil {
		return nil, fmt.Errorf("invalid date/time: %v", err)
	}

	d.id = dt.Format(URLFriendlyDateTimeLayout)
	d.ix = InvalidIndex
	d.DateTimeIn = dt
	d.Data = diveRecord
	return d, nil
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
	return d.DateTimeIn.Add(d.Data.Duration.Value())
}
