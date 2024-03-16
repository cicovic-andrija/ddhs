package main

import (
	"sync"
	"time"
)

const (
	StateWaiting = iota
	StateRunning
	StateFinished
)

var syncJob = NewSyncJob()

type SyncJob struct {
	JobID             string
	state             int
	total             int
	completed         int
	startRunningEvent sync.WaitGroup
}

func NewSyncJob() *SyncJob {
	return &SyncJob{
		JobID: "mockJob",
		state: StateWaiting,
	}
}

func (job *SyncJob) CompletedPct() int {
	if job.state != StateWaiting {
		return 100 * job.completed / job.total // data race is no concern here
	}
	return 0
}

func (job *SyncJob) State() int {
	return job.state
}

func (job *SyncJob) Start() {
	if job.state == StateWaiting {
		job.startRunningEvent.Add(1)
		go job.run()
		job.startRunningEvent.Wait()
	}
}

func (job *SyncJob) run() {
	job.completed = 0
	job.total = 1 // initialize to non-zero
	job.state = StateRunning
	job.startRunningEvent.Done()

	// mock
	job.total = 100
	for job.completed < job.total {
		time.Sleep(100 * time.Millisecond)
		job.completed++
	}

	job.state = StateFinished
}
