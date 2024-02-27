package main

import "time"

// temporary in-memory database
var runs = []Run{
	{
		Num:          1,
		Date:         parseDate("2024-02-22"),
		Location:     "Dunavski kej, Beograd",
		Duration:     parseDuration("36m4s"),
		Distance:     5.18,
		AvgPace:      parseDuration("6m57s"),
		MaxPace:      parseDuration("5m35s"),
		AvgSpeed:     8.6,
		MaxSpeed:     10.7,
		AvgHeartRate: 141,
		MaxHeartRate: 160,
		OneKmLaps: []time.Duration{
			parseDuration("7m57s"),
			parseDuration("7m1s"),
			parseDuration("6m50s"),
			parseDuration("6m20s"),
			parseDuration("6m35s"),
		},
	},
	{
		Num:          2,
		Date:         parseDate("2024-02-27"),
		Location:     "Dunavski kej, Beograd",
		Duration:     parseDuration("36m10s"),
		Distance:     5.5,
		AvgPace:      parseDuration("6m34s"),
		MaxPace:      parseDuration("5m7s"),
		AvgSpeed:     9.1,
		MaxSpeed:     11.6,
		AvgHeartRate: 154,
		MaxHeartRate: 173,
		OneKmLaps: []time.Duration{
			parseDuration("6m56s"),
			parseDuration("6m44s"),
			parseDuration("6m37s"),
			parseDuration("6m25s"),
			parseDuration("6m19s"),
		},
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

type Run struct {
	Num          int
	Date         time.Time
	Location     string
	Duration     time.Duration
	Distance     float32
	AvgPace      time.Duration
	MaxPace      time.Duration
	AvgSpeed     float32
	MaxSpeed     float32
	AvgHeartRate int
	MaxHeartRate int
	OneKmLaps    []time.Duration
}

func (r *Run) ID() int {
	return r.Num - 1
}
