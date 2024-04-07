package main

import (
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
