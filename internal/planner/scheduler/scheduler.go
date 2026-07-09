package scheduler

import "fmt"

type Scheduler interface {
	Schedule(graph any) ([]Task, error)
}

type Task struct {
	ID   string
	Name string
}

type DefaultScheduler struct{}

func NewScheduler() Scheduler {
	return DefaultScheduler{}
}

func (DefaultScheduler) Schedule(graph any) ([]Task, error) {
	if graph == nil {
		return nil, fmt.Errorf("graph is nil")
	}
	return []Task{{ID: "task-1", Name: "bootstrap"}}, nil
}
