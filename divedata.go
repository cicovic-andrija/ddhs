package main

import "time"

// temporary in-memory database
var dives = []*Dive{
	{
		Num:  1,
		Date: parseDate("2024-02-01"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  2,
		Date: parseDate("2024-02-02"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  3,
		Date: parseDate("2024-02-03"),
		Site: "Ada Ciganlija, Beograd",
	},
}

func parseDate(str string) time.Time {
	date, err := time.Parse("2006-01-02", str)
	if err != nil {
		panic(err)
	}
	return date
}

func parseDuration(str string) time.Duration {
	dur, err := time.ParseDuration(str)
	if err != nil {
		panic(err)
	}
	return dur
}

type Dive struct {
	Num  int
	Date time.Time
	Site string
	// Country  string
	// Duration time.Duration
}

func NewDive() *Dive {
	return &Dive{
		Num:  0,
		Date: time.Now().UTC(),
	}
}

func (r *Dive) SetNum(id int) {
	r.Num = id + 1
}

func ntoi(num int) int {
	return num - 1
}

func iton(id int) int {
	return id + 1
}
