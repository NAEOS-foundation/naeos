package scheduler

import (
	"sync"
	"time"
)

// Job

type JobFunc func() error

type Job struct {
	ID         string
	Name       string
	Schedule   *Schedule
	Fn         JobFunc
	LastRun    time.Time
	NextRun    time.Time
	Enabled    bool
	RunCount   int
	ErrorCount int
}

type Schedule struct {
	Interval time.Duration
	Cron     string
}

// Schedule

func Every(duration time.Duration) *Schedule {
	return &Schedule{Interval: duration}
}

func Daily() *Schedule {
	return &Schedule{Interval: 24 * time.Hour}
}

func Hourly() *Schedule {
	return &Schedule{Interval: time.Hour}
}

func Minutely() *Schedule {
	return &Schedule{Interval: time.Minute}
}

// Scheduler

type Scheduler struct {
	jobs     map[string]*Job
	running  bool
	stopCh   chan struct{}
	tickRate time.Duration
	mu       sync.RWMutex
}

func New() *Scheduler {
	return &Scheduler{
		jobs:     make(map[string]*Job),
		stopCh:   make(chan struct{}),
		tickRate: time.Second,
	}
}

func NewWithTickRate(rate time.Duration) *Scheduler {
	return &Scheduler{
		jobs:     make(map[string]*Job),
		stopCh:   make(chan struct{}),
		tickRate: rate,
	}
}

func (s *Scheduler) Add(job *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job.NextRun = time.Now()
	job.Enabled = true
	s.jobs[job.ID] = job
}

func (s *Scheduler) Remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, id)
}

func (s *Scheduler) Get(id string) (*Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[id]
	return job, ok
}

func (s *Scheduler) List() []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

func (s *Scheduler) Enable(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, ok := s.jobs[id]; ok {
		job.Enabled = true
		job.NextRun = time.Now().Add(job.Schedule.Interval)
	}
}

func (s *Scheduler) Disable(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, ok := s.jobs[id]; ok {
		job.Enabled = false
	}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	stopCh := s.stopCh

	go s.run(stopCh)
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	s.running = false
	close(s.stopCh)
	s.stopCh = make(chan struct{})
}

func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *Scheduler) run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(s.tickRate)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *Scheduler) tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for _, job := range s.jobs {
		if !job.Enabled {
			continue
		}

		if now.After(job.NextRun) || now.Equal(job.NextRun) {
			go s.executeJob(job)
			job.LastRun = now
			job.NextRun = now.Add(job.Schedule.Interval)
			job.RunCount++
		}
	}
}

func (s *Scheduler) executeJob(job *Job) {
	if err := job.Fn(); err != nil {
		s.mu.Lock()
		job.ErrorCount++
		s.mu.Unlock()
	}
}

// Job Builder

func NewJob(id, name string, schedule *Schedule, fn JobFunc) *Job {
	return &Job{
		ID:       id,
		Name:     name,
		Schedule: schedule,
		Fn:       fn,
		Enabled:  true,
	}
}

// Stats

type Stats struct {
	TotalJobs   int
	EnabledJobs int
	RunningJobs int
	TotalRuns   int
	TotalErrors int
}

func (s *Scheduler) Stats() *Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &Stats{}
	for _, job := range s.jobs {
		stats.TotalJobs++
		if job.Enabled {
			stats.EnabledJobs++
		}
		stats.TotalRuns += job.RunCount
		stats.TotalErrors += job.ErrorCount
	}
	return stats
}
