package main

import (
	"encoding/json"
	"errors"
	"time"
)

var ErrInvalidDuration = errors.New("invalid duration value")

// Duration is a time.Duration wrapper that can be marshalled (unmarshalled) to (from) JSON.
type Duration struct {
	time.Duration
}

func (d *Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return ErrInvalidDuration
	}
}

func (d *Duration) Value() time.Duration {
	return d.Duration
}
