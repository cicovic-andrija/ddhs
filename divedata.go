package main

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	InvalidIndex = -1

	IDTag   = "id"
	DateTag = "date"
	SiteTag = "site"

	DateLayout = "2006-01-02"
)

type Dive struct {
	id     string
	ix     int
	Date   time.Time `json:"date"`
	TimeIn time.Time `json:"time_in"`
	// TimeOut time.Time `json:"time_out"`
	Site string `json:"site"`
	// Duration time.Duration `json:"duration"`
	// MaxDepth float32 `json:"max_depth"`
	// AvgDepth float32 `json:"avg_depth"`
	// MinWaterTemp float32 `json:"min_water_temp"`
}

func NewID() string {
	id, _ := RandHexString(3)
	return id
}

func NewDive() *Dive {
	return &Dive{
		id:     NewID(),
		ix:     InvalidIndex,
		Date:   time.Now().UTC(),
		TimeIn: time.Now().UTC(),
	}
}

func (d *Dive) ID() string {
	return d.id
}

func (d *Dive) Num() int {
	return d.ix + 1
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

type DiveLog struct {
	sync.RWMutex

	dives      map[string]*Dive
	sorted     DiveList
	renumbered atomic.Bool
}

// The following DiveLog functions are not thread-safe, or they return data pointers
// to the calling function, so they should be used together with embedded locking functions.

func (dl *DiveLog) All() DiveList {
	return dl.sorted
}

func (dl *DiveLog) Find(id string) *Dive {
	return dl.dives[id]
}

func (dl *DiveLog) Insert(dive *Dive) {
	dl.dives[dive.id] = dive
	dl.sorted = append(dl.sorted, dive)
	dive.ix = len(dl.sorted) - 1
	if len(dl.sorted) > 1 {
		for dive.ix > 0 && dive.Date.Before(dl.sorted[dive.ix-1].Date) {
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
	// TODO: figure out a smarter way to do this, including TimeIn
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

var InMemLog DiveLog = DiveLog{
	dives:  make(map[string]*Dive),
	sorted: make(DiveList, 0),
}

func InitInMemLog() {
	parseDate := func(str string) time.Time {
		date, err := time.Parse(DateLayout, str)
		if err != nil {
			panic(err)
		}
		return date
	}

	// parseDuration := func(str string) time.Duration {
	// 	dur, err := time.ParseDuration(str)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	return dur
	// }

	d := parseDate("2023-12-01")
	i := 10
	for i > 0 {
		dive := NewDive()
		dive.Site = "Alif Alif Atoll, Maldives"
		dive.Date = d.AddDate(0, 0, 51-i)
		InMemLog.Insert(dive)
		i--
	}
}
