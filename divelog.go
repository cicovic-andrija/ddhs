package main

import "time"

const (
	NumTag  = "num"
	DateTag = "date"
	SiteTag = "site"

	DateLayout = "2006-01-02"
)

type Dive struct {
	Num  int       `json:"num"`
	Date time.Time `json:"date"`
	Site string    `json:"site"`
	// Duration time.Duration `json:"duration"`
	// MaxDepth float32 `json:"max_depth"`
	// AvgDepth float32 `json:"avg_depth"`
	// MinWaterTemp float32 `json:"min_water_temp"`
}

type DiveLog []*Dive

// sensible empty values
func NewDive() *Dive {
	return &Dive{
		Num:  0,
		Date: time.Now().UTC(),
	}
}

func ntoi(num int) int {
	return num - 1
}

func iton(id int) int {
	return id + 1
}

func (dl DiveLog) Filter(predicate func(*Dive) bool) DiveLog {
	filtered := make([]*Dive, 0, len(dl))
	for _, dive := range dl {
		if predicate(dive) {
			filtered = append(filtered, dive)
		}
	}
	return filtered
}

// ---------------------------------------------------------------------------------------------------------------------

func parseDate(str string) time.Time {
	date, err := time.Parse(DateLayout, str)
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

// temporary in-memory database
var dives = DiveLog{
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
	{
		Num:  4,
		Date: parseDate("2024-02-04"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  5,
		Date: parseDate("2024-02-05"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  6,
		Date: parseDate("2024-02-06"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  7,
		Date: parseDate("2024-02-07"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  8,
		Date: parseDate("2024-02-08"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  9,
		Date: parseDate("2024-02-09"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  10,
		Date: parseDate("2024-02-10"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  11,
		Date: parseDate("2024-02-11"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  12,
		Date: parseDate("2024-02-12"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  13,
		Date: parseDate("2024-02-13"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  14,
		Date: parseDate("2024-02-14"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  15,
		Date: parseDate("2024-02-15"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  16,
		Date: parseDate("2024-02-16"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  17,
		Date: parseDate("2024-02-17"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  18,
		Date: parseDate("2024-02-18"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  19,
		Date: parseDate("2024-02-19"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  20,
		Date: parseDate("2024-02-20"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  21,
		Date: parseDate("2024-02-21"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  22,
		Date: parseDate("2024-02-22"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  23,
		Date: parseDate("2024-02-23"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  24,
		Date: parseDate("2024-02-24"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  25,
		Date: parseDate("2024-02-25"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  26,
		Date: parseDate("2024-02-26"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  27,
		Date: parseDate("2024-02-27"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  28,
		Date: parseDate("2024-02-28"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  29,
		Date: parseDate("2024-02-29"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  30,
		Date: parseDate("2024-03-01"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  31,
		Date: parseDate("2024-03-02"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  32,
		Date: parseDate("2024-03-03"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  33,
		Date: parseDate("2024-03-04"),
		Site: "Ada Ciganlija, Beograd",
	},
	{
		Num:  34,
		Date: parseDate("2024-03-05"),
		Site: "Ada Ciganlija, Beograd",
	},
}
