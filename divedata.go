package main

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	InvalidIndex = -1

	IDTag       = "id"
	SiteTag     = "site"
	DateTag     = "date"
	TimeInTag   = "time_in"
	DurationTag = "duration"

	TimeLayout                = "15:04"
	DateLayout                = "2006-01-02"
	DateTimeLayout            = "2006-01-02T15:04"
	URLFriendlyDateTimeLayout = "2006-01-02T15-04"
)

type Dive struct {
	id            string
	ix            int
	DateTimeIn    time.Time     `json:"-"`
	DateTimeInStr string        `json:"date_time_in"`
	Site          string        `json:"site"`
	Duration      time.Duration `json:"duration"`
}

// NewDive returns a pointer to a new Dive struct, with all fields initialized to their "zero" or default/invalid value.
func NewDive() *Dive {
	return &Dive{ix: InvalidIndex}
}

func (d *Dive) SetDateTimeInAndAssignID(new time.Time) {
	if d.id == "" {
		d.DateTimeIn = new
		d.DateTimeInStr = d.DateTimeIn.Format(DateTimeLayout)
		d.id = d.DateTimeIn.Format(URLFriendlyDateTimeLayout)
	}
}

func (d *Dive) ID() string {
	return d.id
}

func (d *Dive) Num() int {
	return d.ix + 1
}

func (d *Dive) TimeOut() time.Time {
	return d.DateTimeIn.Add(d.Duration)
}

// Reconstruct the complete state based on partially initialized set of fields.
// TODO: Additional validation?
func (d *Dive) Reconstruct() error {
	dti, err := time.Parse(DateTimeLayout, d.DateTimeInStr)
	if err != nil {
		return fmt.Errorf("invalid date time in: %v", err)
	}
	d.SetDateTimeInAndAssignID(dti)
	return nil
}

type DiveList []*Dive

func (s DiveList) Filter(predicate func(*Dive) bool) []*Dive {
	filtered := make([]*Dive, 0, len(s))
	for _, dive := range s {
		if predicate(dive) {
			filtered = append(filtered, dive)
		}
	}
	return filtered
}

// Reconstruct all dives.
func (s DiveList) Reconstruct() error {
	for i, d := range s {
		if err := d.Reconstruct(); err != nil {
			return generateReconstructionError(err.Error(), fmt.Sprintf("/dives/%d", i))
		}
	}
	return nil
}

// DiveLog doesn't have thread-safe functions, and defines functions that return data pointers.
// Callers are responsible for read/write locking.
type DiveLog struct {
	sync.RWMutex

	dives         map[string]*Dive //
	sorted        DiveList         //
	renumbered    atomic.Bool      //
	sequence      uint64           // persistence: always one ahead from persistent storage, updated on save
	lastPersisted time.Time        // persistence: read from persistent storage, updated on save
}

func (dl *DiveLog) All() DiveList {
	return dl.sorted
}

func (dl *DiveLog) Find(id string) *Dive {
	return dl.dives[id]
}

// Reconstruct `dives` from a DiveList. Also, make sure `sorted` is initialized with a sorted copy of the DiveList.
func (dl *DiveLog) Reconstruct(dives DiveList) error {
	err := dives.Reconstruct()
	if err != nil {
		return err
	}

	// Ideally, no errors after this point because internal state will be changed.

	sort.Slice(dives, func(i int, j int) bool { return dives[i].DateTimeIn.Before(dives[j].DateTimeIn) })
	dl.sorted = dives

	for ix, dive := range dl.sorted {
		dive.ix = ix
		dl.dives[dive.id] = dive
	}

	dl.renumbered.Store(false)

	return nil
}

func (dl *DiveLog) Insert(dive *Dive) {
	dl.dives[dive.id] = dive
	dl.sorted = append(dl.sorted, dive)
	dive.ix = len(dl.sorted) - 1
	if len(dl.sorted) > 1 {
		for dive.ix > 0 && dive.DateTimeIn.Before(dl.sorted[dive.ix-1].DateTimeIn) {
			dl.sorted[dive.ix] = dl.sorted[dive.ix-1]
			dl.sorted[dive.ix].ix++
			dl.sorted[dive.ix-1] = dive
			dive.ix--
		}
	}
	if dive.ix < len(dl.sorted)-1 {
		dl.renumbered.Store(true)
	}
}

func (dl *DiveLog) Replace(existing *Dive, new *Dive) {
	new.id = existing.id
	// TODO: Figure out a smarter way to do this, including TimeIn.
	dl.Delete(existing.id)
	dl.Insert(new)
}

func (dl *DiveLog) Delete(id string) (found bool) {
	var (
		dive *Dive
	)

	dive, found = dl.dives[id]
	if !found {
		return
	}
	delete(dl.dives, id)

	if dive.ix < len(dl.sorted)-1 {
		for i := dive.ix; i < len(dl.sorted)-1; i++ {
			dl.sorted[i] = dl.sorted[i+1]
			dl.sorted[i].ix = i
		}
		dl.renumbered.Store(true)
	}
	dl.sorted = dl.sorted[:len(dl.sorted)-1]

	return
}

func (dl *DiveLog) IsRenumbered() bool {
	return dl.renumbered.CompareAndSwap(true, false)
}

func generateReconstructionError(reason string, hint string) error {
	return fmt.Errorf("reconstruction failed: %s @ %s", reason, hint)
}
