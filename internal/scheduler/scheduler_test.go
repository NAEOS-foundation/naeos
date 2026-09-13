// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("expected scheduler to be created")
	}
}

func TestNewWithTickRate(t *testing.T) {
	s := NewWithTickRate(10 * time.Millisecond)
	if s == nil {
		t.Fatal("expected scheduler to be created")
	}
}

func TestAddJob(t *testing.T) {
	s := New()
	job := NewJob("job1", "Test Job", Every(time.Second), func() error {
		return nil
	})

	s.Add(job)

	got, ok := s.Get("job1")
	if !ok {
		t.Fatal("expected job to be found")
	}
	if got.Name != "Test Job" {
		t.Errorf("expected name 'Test Job', got %s", got.Name)
	}
}

func TestRemoveJob(t *testing.T) {
	s := New()
	job := NewJob("job1", "Test Job", Every(time.Second), func() error {
		return nil
	})

	s.Add(job)
	s.Remove("job1")

	_, ok := s.Get("job1")
	if ok {
		t.Error("expected job to be removed")
	}
}

func TestListJobs(t *testing.T) {
	s := New()
	s.Add(NewJob("job1", "Job 1", Every(time.Second), func() error { return nil }))
	s.Add(NewJob("job2", "Job 2", Every(time.Second), func() error { return nil }))

	jobs := s.List()
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestEnableDisable(t *testing.T) {
	s := New()
	job := NewJob("job1", "Test Job", Every(time.Second), func() error {
		return nil
	})

	s.Add(job)
	s.Disable("job1")

	got, _ := s.Get("job1")
	if got.Enabled {
		t.Error("expected job to be disabled")
	}

	s.Enable("job1")
	got, _ = s.Get("job1")
	if !got.Enabled {
		t.Error("expected job to be enabled")
	}
}

func TestStartStop(t *testing.T) {
	s := NewWithTickRate(10 * time.Millisecond)

	s.Start()
	if !s.IsRunning() {
		t.Error("expected scheduler to be running")
	}

	s.Stop()
	if s.IsRunning() {
		t.Error("expected scheduler to be stopped")
	}
}

func TestExecuteJob(t *testing.T) {
	s := NewWithTickRate(10 * time.Millisecond)

	var count int32
	job := NewJob("job1", "Test Job", Every(10*time.Millisecond), func() error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	s.Add(job)
	s.Start()

	time.Sleep(100 * time.Millisecond)
	s.Stop()

	if atomic.LoadInt32(&count) == 0 {
		t.Error("expected job to be executed")
	}
}

func TestJobExecution(t *testing.T) {
	var executed bool
	job := NewJob("job1", "Test Job", Every(time.Second), func() error {
		executed = true
		return nil
	})

	job.Fn()

	if !executed {
		t.Error("expected job function to be executed")
	}
}

func TestScheduleIntervals(t *testing.T) {
	tests := []struct {
		name     string
		schedule *Schedule
		expected time.Duration
	}{
		{"every second", Every(time.Second), time.Second},
		{"every minute", Every(time.Minute), time.Minute},
		{"daily", Daily(), 24 * time.Hour},
		{"hourly", Hourly(), time.Hour},
		{"minutely", Minutely(), time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.schedule.Interval != tt.expected {
				t.Errorf("expected interval %v, got %v", tt.expected, tt.schedule.Interval)
			}
		})
	}
}

func TestStats(t *testing.T) {
	s := New()

	s.Add(NewJob("job1", "Job 1", Every(time.Second), func() error { return nil }))
	s.Add(NewJob("job2", "Job 2", Every(time.Second), func() error { return nil }))
	s.Disable("job2")

	stats := s.Stats()
	if stats.TotalJobs != 2 {
		t.Errorf("expected 2 total jobs, got %d", stats.TotalJobs)
	}
	if stats.EnabledJobs != 1 {
		t.Errorf("expected 1 enabled job, got %d", stats.EnabledJobs)
	}
}

func TestJobDefaults(t *testing.T) {
	job := NewJob("job1", "Test", Every(time.Second), func() error { return nil })

	if !job.Enabled {
		t.Error("expected job to be enabled by default")
	}

	if job.RunCount != 0 {
		t.Errorf("expected run count 0, got %d", job.RunCount)
	}

	if job.ErrorCount != 0 {
		t.Errorf("expected error count 0, got %d", job.ErrorCount)
	}
}

func TestMultipleJobs(t *testing.T) {
	s := NewWithTickRate(10 * time.Millisecond)

	var count1, count2 int32

	s.Add(NewJob("job1", "Job 1", Every(10*time.Millisecond), func() error {
		atomic.AddInt32(&count1, 1)
		return nil
	}))

	s.Add(NewJob("job2", "Job 2", Every(10*time.Millisecond), func() error {
		atomic.AddInt32(&count2, 1)
		return nil
	}))

	s.Start()
	time.Sleep(100 * time.Millisecond)
	s.Stop()

	if atomic.LoadInt32(&count1) == 0 {
		t.Error("expected job1 to be executed")
	}
	if atomic.LoadInt32(&count2) == 0 {
		t.Error("expected job2 to be executed")
	}
}
