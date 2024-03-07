package main

import "sync"

type DiveLog2 struct {
	dives map[string]*Dive
	index []*Dive
	mu    sync.RWMutex
}

func (diveLog *DiveLog2) Find(id string) *Dive {
	return nil
}

func (diveLog *DiveLog2) Insert(dive *Dive) {

}
