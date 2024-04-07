package main

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
