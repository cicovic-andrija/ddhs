package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cicovic-andrija/libgo/logging"
)

const (
	DataDirectory       = "data"
	DiveLogFileName     = "divelog.json"
	TempDiveLogFileName = "divelog.tmp.json"
)

type PersistedDiveLog struct {
	Version  string   `json:"version"`
	Modified string   `json:"modified"`
	Dives    DiveList `json:"dives"`
}

func saveAsync(memlog *DiveLog) {
	memlog.Lock()
	if err := memlog.save(); err != nil {
		traceServerMessage(logging.SevError, "persistence of log sequence %d failed: %v", memlog.sequence, err)
	}
	memlog.Unlock()
}

func (dl *DiveLog) save() error {
	tmpFile, err := os.OpenFile(
		filepath.Join(DataDirectory, TempDiveLogFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("temp log file creation failed: %v", err)
	}
	defer os.Remove(TempDiveLogFileName)
	defer tmpFile.Close()

	modifiedTime := time.Now().UTC()
	persistedLog := &PersistedDiveLog{
		Version:  fmt.Sprintf("%d:%d", 1, dl.sequence),
		Modified: modifiedTime.Format(time.RFC3339),
		Dives:    dl.sorted,
	}
	if err = json.NewEncoder(tmpFile).Encode(persistedLog); err != nil {
		return fmt.Errorf("encode log operation failed: %v", err)
	}
	if err = os.Rename(
		filepath.Join(DataDirectory, TempDiveLogFileName),
		filepath.Join(DataDirectory, DiveLogFileName),
	); err != nil {
		return fmt.Errorf("write log operation failed: %v", err)
	}

	dl.sequence++
	dl.lastPersisted = modifiedTime
	return nil
}
