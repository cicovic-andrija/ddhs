package main

import (
	"testing"
	"time"
)

func TestDiveLogOps(t *testing.T) {
	diveLog := NewDiveLog()

	diveA := NewDive(datetime("2023-04-03T10:30"))
	diveLog.Insert(diveA)

	diveB := NewDive(datetime("2023-04-04T10:00"))
	diveLog.Insert(diveB)

	if got := diveLog.sorted[0]; got != diveA {
		t.Errorf("sorted[0]: got %s, want %s", got.ID(), diveA.ID())
	}

	if got := diveLog.sorted[1]; got != diveB {
		t.Errorf("sorted[1]: got %s, want %s", got.ID(), diveB.ID())
	}

	if got := diveLog.IsRenumbered(); got != false {
		t.Errorf("IsRenumbered: got %t, want %t", got, false)
	}

	diveC := NewDive(datetime("2023-04-03T13:05"))
	diveLog.Insert(diveC)

	if got := diveLog.sorted[1]; got != diveC {
		t.Errorf("sorted[1]: got %s, want %s", got.ID(), diveC.ID())
	}

	if got := diveLog.sorted[2]; got != diveB {
		t.Errorf("sorted[2]: got %s, want %s", got.ID(), diveB.ID())
	}

	if got := diveLog.IsRenumbered(); got != true {
		t.Errorf("IsRenumbered: got %t, want %t", got, true)
	}
}

func datetime(str string) time.Time {
	if dt, err := time.Parse(DateTimeLayout, str); err != nil {
		panic(err)
	} else {
		return dt
	}
}
