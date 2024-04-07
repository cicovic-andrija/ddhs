package main

import "fmt"

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
			return fmt.Errorf("reconstruction failed: %v @ %s", err, fmt.Sprintf("/dives/%d", i))
		}
	}
	return nil
}
