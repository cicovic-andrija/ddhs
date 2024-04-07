package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cicovic-andrija/libgo/logging"
)

// Terminology:
// mlog - memory log
// plog - persisted log

const (
	DataDirectory       = "data"
	DiveLogFileName     = "divelog.json"
	TempDiveLogFileName = "divelog.tmp.json"

	LogMajor = 1
)

var ErrCorruptedLog = errors.New("corrupted log file")

var MLog DiveLog = DiveLog{
	dives:  make(map[string]*Dive),
	sorted: make(DiveList, 0),
}

type PersistedDiveLog struct {
	Version  string        `json:"version"`
	Modified string        `json:"modified"`
	Dives    []*DiveRecord `json:"dives"`
}

func ensureDataLoadAsync() {
	MLog.Lock()
	if err := MLog.load(); err != nil {
		crash("load dive data operation failed: %v", err)
	}
	MLog.Unlock()
}

func (mlog *DiveLog) load() error {
	var (
		modified time.Time
		sequence uint64
	)

	plogFile, err := os.Open(filepath.Join(DataDirectory, DiveLogFileName))
	if err != nil {
		return fmt.Errorf("read log operation failed: %v", err)
	}

	plog := &PersistedDiveLog{}
	if err = json.NewDecoder(plogFile).Decode(plog); err != nil {
		return fmt.Errorf("decode log operation failed: %v", err)
	}

	// First, validate the "header".
	if parts := strings.Split(plog.Version, ":"); len(parts) != 2 {
		return ErrCorruptedLog
	} else {
		if maj, err := strconv.Atoi(parts[0]); err != nil || maj != LogMajor {
			return ErrCorruptedLog
		}
		if seq, err := strconv.Atoi(parts[1]); err != nil {
			return ErrCorruptedLog
		} else {
			sequence = uint64(seq)
		}
	}
	if mod, err := time.Parse(time.RFC3339, plog.Modified); err != nil {
		return ErrCorruptedLog
	} else {
		modified = mod
	}

	// Second, extract and reconstruct dive data.
	if err = mlog.Reconstruct(plog.Dives); err == nil {
		mlog.sequence = sequence + 1
		mlog.lastPersisted = modified
	}
	return err
}

func saveAsync(mlog *DiveLog) {
	mlog.Lock()
	if err := mlog.save(); err != nil {
		trace(logging.SevError, "persistence of log sequence %d failed: %v", mlog.sequence, err)
	}
	mlog.Unlock()
}

func (dl *DiveLog) save() error {
	tmpFile, err := os.OpenFile(
		filepath.Join(DataDirectory, TempDiveLogFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("temp log file creation failed: %v", err)
	}
	defer os.Remove(TempDiveLogFileName)
	defer tmpFile.Close()

	diveRecords := make([]*DiveRecord, 0, len(dl.sorted))
	for _, dive := range dl.sorted {
		diveRecords = append(diveRecords, dive.Data)
	}
	modifiedTime := time.Now().UTC()
	plog := &PersistedDiveLog{
		Version:  fmt.Sprintf("%d:%d", LogMajor, dl.sequence),
		Modified: modifiedTime.Format(time.RFC3339),
		Dives:    diveRecords,
	}
	if err = json.NewEncoder(tmpFile).Encode(plog); err != nil {
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
