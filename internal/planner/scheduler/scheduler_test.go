// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"strings"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
)

// --- Existing tests (preserved) ---

func TestSchedulerReturnsErrorForNilInput(t *testing.T) {
	s := NewScheduler()
	_, err := s.Schedule(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestSchedulerGeneratesTasksFromNEIR(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "acme-api"},
		Modules: []module.Module{
			{Name: "auth", Path: "./internal/auth"},
			{Name: "user", Path: "./internal/user"},
		},
		Services: []service.Service{
			{Name: "api", Kind: "http", Port: 8080},
		},
	}

	s := NewScheduler()
	tasks, err := s.Schedule(neir)
	if err != nil {
		t.Fatalf("Schedule returned error: %v", err)
	}

	if len(tasks) < 5 {
		t.Fatalf("expected at least 5 tasks, got %d", len(tasks))
	}

	if tasks[0].Name != "validate-specification" {
		t.Fatalf("expected first task validate-specification, got %s", tasks[0].Name)
	}

	if tasks[1].Name != "build-model" {
		t.Fatalf("expected second task build-model, got %s", tasks[1].Name)
	}

	foundAuth := false
	foundUser := false
	for _, task := range tasks {
		if task.Name == "generate-module-auth" {
			foundAuth = true
		}
		if task.Name == "generate-module-user" {
			foundUser = true
		}
	}
	if !foundAuth || !foundUser {
		t.Fatalf("expected module tasks for auth and user, got tasks: %v", tasks)
	}

	lastTask := tasks[len(tasks)-1]
	if lastTask.Name != "validate-output" {
		t.Fatalf("expected last task validate-output, got %s", lastTask.Name)
	}
}

func TestSchedulerWithoutServices(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "simple-app"},
		Modules: []module.Module{
			{Name: "core", Path: "./internal/core"},
		},
	}

	s := NewScheduler()
	tasks, err := s.Schedule(neir)
	if err != nil {
		t.Fatalf("Schedule returned error: %v", err)
	}

	if len(tasks) < 4 {
		t.Fatalf("expected at least 4 tasks, got %d", len(tasks))
	}

	hasServiceTask := false
	for _, task := range tasks {
		if task.Name == "generate-service-api" {
			hasServiceTask = true
		}
	}
	if hasServiceTask {
		t.Fatal("should not have service task when no services defined")
	}
}

func TestSchedulerFallbackForNonNEIRInput(t *testing.T) {
	s := NewScheduler()
	tasks, err := s.Schedule("not-a-neir")
	if err != nil {
		t.Fatalf("Schedule returned error: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Name != "bootstrap" {
		t.Fatalf("expected fallback bootstrap task, got %v", tasks)
	}
}

func TestParallelGroups(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "acme-api"},
		Modules: []module.Module{
			{Name: "auth", Path: "./internal/auth"},
			{Name: "user", Path: "./internal/user"},
		},
		Services: []service.Service{
			{Name: "api", Kind: "http", Port: 8080},
		},
	}

	s := NewScheduler()
	tasks, err := s.Schedule(neir)
	if err != nil {
		t.Fatalf("Schedule returned error: %v", err)
	}

	groups := s.ParallelGroups(tasks)
	if len(groups) == 0 {
		t.Fatal("expected at least one parallel group")
	}

	for _, g := range groups {
		if len(g.Tasks) == 0 {
			t.Fatalf("group at level %d has no tasks", g.Level)
		}
		for _, task := range g.Tasks {
			if task.Priority != g.Level {
				t.Fatalf("task %s has priority %d but is in group level %d", task.Name, task.Priority, g.Level)
			}
		}
	}

	moduleGroups := 0
	for _, g := range groups {
		if g.Level == 2 {
			moduleGroups++
			if len(g.Tasks) != 2 {
				t.Fatalf("expected 2 module tasks in parallel group, got %d", len(g.Tasks))
			}
		}
	}
	if moduleGroups != 1 {
		t.Fatalf("expected 1 group at level 2, got %d", moduleGroups)
	}
}

func TestParallelGroupsEmpty(t *testing.T) {
	s := NewScheduler()
	groups := s.ParallelGroups(nil)
	if groups != nil {
		t.Fatal("expected nil for nil input")
	}

	groups = s.ParallelGroups([]Task{})
	if groups != nil {
		t.Fatal("expected nil for empty input")
	}
}

// --- DAGScheduler tests ---

func TestDAGSchedulerNilInput(t *testing.T) {
	ds := NewDAGScheduler()
	_, err := ds.Schedule(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestDAGSchedulerFallbackNonNEIR(t *testing.T) {
	ds := NewDAGScheduler()
	tasks, err := ds.Schedule("not-a-neir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Name != "bootstrap" {
		t.Fatalf("expected bootstrap task, got %v", tasks)
	}
}

func TestDAGSchedulerRespectsDeps(t *testing.T) {
	ds := NewDAGScheduler()
	input := []Task{
		{ID: "a", Name: "a", Priority: 2},
		{ID: "b", Name: "b", Dependencies: []string{"a"}, Priority: 1},
		{ID: "c", Name: "c", Dependencies: []string{"b"}},
	}
	result := ds.scheduleFromTasks(input)

	index := make(map[string]int)
	for i, task := range result {
		index[task.ID] = i
	}
	if index["a"] >= index["b"] {
		t.Fatalf("a should come before b, got indices a=%d b=%d", index["a"], index["b"])
	}
	if index["b"] >= index["c"] {
		t.Fatalf("b should come before c, got indices b=%d c=%d", index["b"], index["c"])
	}
}

func TestDAGSchedulerParallelGroupsDAG(t *testing.T) {
	ds := NewDAGScheduler()
	input := []Task{
		{ID: "a", Name: "a"},
		{ID: "b", Name: "b"},
		{ID: "c", Name: "c", Dependencies: []string{"a", "b"}},
	}
	groups := ds.ParallelGroups(input)
	if len(groups) != 2 {
		t.Fatalf("expected 2 levels, got %d", len(groups))
	}
	if len(groups[0].Tasks) != 2 {
		t.Fatalf("expected 2 tasks in first level, got %d", len(groups[0].Tasks))
	}
	if len(groups[1].Tasks) != 1 {
		t.Fatalf("expected 1 task in second level, got %d", len(groups[1].Tasks))
	}
}

func TestDAGSchedulerSingleTask(t *testing.T) {
	ds := NewDAGScheduler()
	input := []Task{{ID: "x", Name: "x"}}
	result := ds.scheduleFromTasks(input)
	if len(result) != 1 || result[0].ID != "x" {
		t.Fatalf("expected single task x, got %v", result)
	}
}

func TestDAGSchedulerWithNEIR(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{
			{Name: "m1", Path: "./m1"},
		},
		Services: []service.Service{
			{Name: "svc1", Kind: "http"},
		},
	}
	ds := NewDAGScheduler()
	tasks, err := ds.Schedule(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) < 5 {
		t.Fatalf("expected at least 5 tasks, got %d", len(tasks))
	}
}

// --- CriticalPath tests ---

func TestCriticalPathEmpty(t *testing.T) {
	path, err := CriticalPath(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != nil {
		t.Fatalf("expected nil, got %v", path)
	}
}

func TestCriticalPathSingleTask(t *testing.T) {
	tasks := []Task{{ID: "a", Name: "a", Duration: 5 * time.Second}}
	path, err := CriticalPath(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) != 1 || path[0].ID != "a" {
		t.Fatalf("expected [a], got %v", path)
	}
}

func TestCriticalPathLinearChain(t *testing.T) {
	tasks := []Task{
		{ID: "a", Name: "a", Duration: 1 * time.Second},
		{ID: "b", Name: "b", Dependencies: []string{"a"}, Duration: 3 * time.Second},
		{ID: "c", Name: "c", Dependencies: []string{"b"}, Duration: 2 * time.Second},
	}
	path, err := CriticalPath(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) != 3 {
		t.Fatalf("expected path length 3, got %d", len(path))
	}
	ids := make([]string, len(path))
	for i, t := range path {
		ids[i] = t.ID
	}
	if ids[0] != "a" || ids[1] != "b" || ids[2] != "c" {
		t.Fatalf("expected [a b c], got %v", ids)
	}
}

func TestCriticalPathSelectsLongestBranch(t *testing.T) {
	tasks := []Task{
		{ID: "a", Name: "a", Duration: 1 * time.Second},
		{ID: "b", Name: "b", Dependencies: []string{"a"}, Duration: 10 * time.Second},
		{ID: "c", Name: "c", Dependencies: []string{"a"}, Duration: 2 * time.Second},
		{ID: "d", Name: "d", Dependencies: []string{"b"}, Duration: 1 * time.Second},
		{ID: "e", Name: "e", Dependencies: []string{"c"}, Duration: 1 * time.Second},
	}
	path, err := CriticalPath(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ids := make([]string, len(path))
	for i, t := range path {
		ids[i] = t.ID
	}
	if ids[len(ids)-1] != "d" {
		t.Fatalf("expected critical path to end at d, got %v", ids)
	}
}

func TestCriticalPathRejectsInvalidDeps(t *testing.T) {
	tasks := []Task{
		{ID: "a", Name: "a", Dependencies: []string{"nonexistent"}},
	}
	_, err := CriticalPath(tasks)
	if err == nil {
		t.Fatal("expected error for missing dependency")
	}
}

func TestEstimateDurationEmpty(t *testing.T) {
	d := EstimateDuration(nil)
	if d != 0 {
		t.Fatalf("expected 0, got %v", d)
	}
}

func TestEstimateDurationSingleTask(t *testing.T) {
	tasks := []Task{{ID: "a", Duration: 42 * time.Second}}
	d := EstimateDuration(tasks)
	if d != 42*time.Second {
		t.Fatalf("expected 42s, got %v", d)
	}
}

func TestEstimateDurationCriticalPath(t *testing.T) {
	tasks := []Task{
		{ID: "a", Duration: 2 * time.Second},
		{ID: "b", Dependencies: []string{"a"}, Duration: 5 * time.Second},
		{ID: "c", Dependencies: []string{"a"}, Duration: 1 * time.Second},
		{ID: "d", Dependencies: []string{"b"}, Duration: 3 * time.Second},
	}
	d := EstimateDuration(tasks)
	if d != 10*time.Second {
		t.Fatalf("expected 10s, got %v", d)
	}
}

// --- ValidateSchedule tests ---

func TestValidateScheduleValid(t *testing.T) {
	tasks := []Task{
		{ID: "a", Dependencies: []string{}},
		{ID: "b", Dependencies: []string{"a"}},
	}
	if err := ValidateSchedule(tasks); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateScheduleDuplicateID(t *testing.T) {
	tasks := []Task{
		{ID: "a"},
		{ID: "a"},
	}
	if err := ValidateSchedule(tasks); err == nil {
		t.Fatal("expected error for duplicate ID")
	}
}

func TestValidateScheduleMissingDep(t *testing.T) {
	tasks := []Task{
		{ID: "a", Dependencies: []string{"z"}},
	}
	if err := ValidateSchedule(tasks); err == nil {
		t.Fatal("expected error for missing dependency")
	}
}

func TestValidateScheduleCycle(t *testing.T) {
	tasks := []Task{
		{ID: "a", Dependencies: []string{"b"}},
		{ID: "b", Dependencies: []string{"a"}},
	}
	if err := ValidateSchedule(tasks); err == nil {
		t.Fatal("expected error for cycle")
	}
}

func TestValidateScheduleLongCycle(t *testing.T) {
	tasks := []Task{
		{ID: "a", Dependencies: []string{"c"}},
		{ID: "b", Dependencies: []string{"a"}},
		{ID: "c", Dependencies: []string{"b"}},
	}
	if err := ValidateSchedule(tasks); err == nil {
		t.Fatal("expected error for 3-node cycle")
	}
}

func TestValidateScheduleEmpty(t *testing.T) {
	if err := ValidateSchedule(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- ScheduleStrategy tests ---

func TestPriorityStrategy(t *testing.T) {
	ps := PriorityStrategy{}
	tasks := []Task{
		{ID: "a", Priority: 3},
		{ID: "b", Priority: 1},
		{ID: "c", Priority: 2},
		{ID: "d", Priority: 1},
	}
	groups := ps.Apply(tasks)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if len(groups[0]) != 2 {
		t.Fatalf("expected 2 tasks in priority 1 group, got %d", len(groups[0]))
	}
	if groups[0][0].ID != "b" || groups[0][1].ID != "d" {
		t.Fatalf("expected [b d] in priority 1, got %v", groups[0])
	}
}

func TestPriorityStrategyEmpty(t *testing.T) {
	ps := PriorityStrategy{}
	groups := ps.Apply(nil)
	if groups != nil {
		t.Fatal("expected nil")
	}
}

func TestRoundRobinStrategy(t *testing.T) {
	rr := RoundRobinStrategy{}
	tasks := []Task{
		{ID: "c"},
		{ID: "a"},
		{ID: "b"},
	}
	groups := rr.Apply(tasks)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if groups[0][0].ID != "a" || groups[1][0].ID != "b" || groups[2][0].ID != "c" {
		t.Fatalf("expected sorted order a,b,c, got %v %v %v", groups[0][0].ID, groups[1][0].ID, groups[2][0].ID)
	}
}

func TestRoundRobinStrategyEmpty(t *testing.T) {
	rr := RoundRobinStrategy{}
	groups := rr.Apply(nil)
	if groups != nil {
		t.Fatal("expected nil")
	}
}

func TestResourceAwareStrategyGroupsCorrectly(t *testing.T) {
	ra := ResourceAwareStrategy{Capacity: 5}
	tasks := []Task{
		{ID: "a", Weight: 3, Priority: 0},
		{ID: "b", Weight: 2, Priority: 0},
		{ID: "c", Weight: 1, Priority: 0},
	}
	groups := ra.Apply(tasks)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if len(groups[0]) != 2 || groups[0][0].Weight+groups[0][1].Weight != 5 {
		t.Fatalf("expected first group weight=5, got %v", groups[0])
	}
}

func TestResourceAwareStrategyDefaultWeight(t *testing.T) {
	ra := ResourceAwareStrategy{Capacity: 2}
	tasks := []Task{
		{ID: "a", Weight: 0},
		{ID: "b", Weight: 0},
	}
	groups := ra.Apply(tasks)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups (each weight=1), got %d", len(groups))
	}
}

func TestResourceAwareStrategyEmpty(t *testing.T) {
	ra := ResourceAwareStrategy{Capacity: 5}
	groups := ra.Apply(nil)
	if groups != nil {
		t.Fatal("expected nil")
	}
}

func TestResourceAwareStrategyZeroCapacity(t *testing.T) {
	ra := ResourceAwareStrategy{Capacity: 0}
	tasks := []Task{{ID: "a"}}
	groups := ra.Apply(tasks)
	if groups != nil {
		t.Fatal("expected nil for zero capacity")
	}
}

// --- TaskGroup tests ---

func TestTaskGroupCreation(t *testing.T) {
	tasks := []Task{
		{ID: "a", Name: "a", Duration: 2 * time.Second, Weight: 3},
		{ID: "b", Name: "b", Duration: 3 * time.Second, Weight: 2},
	}
	tg := NewTaskGroup("build", tasks)
	if tg.Name != "build" {
		t.Fatalf("expected name 'build', got %s", tg.Name)
	}
	if len(tg.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tg.Tasks))
	}
}

func TestTaskGroupTaskIDs(t *testing.T) {
	tasks := []Task{
		{ID: "x"},
		{ID: "y"},
		{ID: "z"},
	}
	tg := NewTaskGroup("test", tasks)
	ids := tg.TaskIDs()
	if len(ids) != 3 || ids[0] != "x" || ids[1] != "y" || ids[2] != "z" {
		t.Fatalf("expected [x y z], got %v", ids)
	}
}

func TestTaskGroupTotalDuration(t *testing.T) {
	tasks := []Task{
		{Duration: 5 * time.Second},
		{Duration: 10 * time.Second},
	}
	tg := NewTaskGroup("g", tasks)
	if tg.TotalDuration() != 15*time.Second {
		t.Fatalf("expected 15s, got %v", tg.TotalDuration())
	}
}

func TestTaskGroupTotalWeight(t *testing.T) {
	tasks := []Task{
		{Weight: 3},
		{Weight: 7},
	}
	tg := NewTaskGroup("g", tasks)
	if tg.TotalWeight() != 10 {
		t.Fatalf("expected 10, got %d", tg.TotalWeight())
	}
}

// --- VisualizeSchedule tests ---

func TestVisualizeScheduleEmpty(t *testing.T) {
	result := VisualizeSchedule(nil)
	if result != "" {
		t.Fatalf("expected empty, got %q", result)
	}
}

func TestVisualizeScheduleNonEmpty(t *testing.T) {
	tasks := []Task{
		{ID: "a", Name: "build", Priority: 0, Duration: 2 * time.Second},
		{ID: "b", Name: "test", Priority: 1, Duration: 3 * time.Second},
	}
	result := VisualizeSchedule(tasks)
	if !strings.Contains(result, "build") {
		t.Fatal("expected 'build' in visualization")
	}
	if !strings.Contains(result, "test") {
		t.Fatal("expected 'test' in visualization")
	}
	if !strings.Contains(result, "#") {
		t.Fatal("expected '#' bar characters in visualization")
	}
}

func TestVisualizeScheduleShowsPriorities(t *testing.T) {
	tasks := []Task{
		{ID: "a", Name: "first", Priority: 0},
		{ID: "b", Name: "second", Priority: 5},
	}
	result := VisualizeSchedule(tasks)
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines (header + sep + 2 tasks), got %d", len(lines))
	}
	if !strings.Contains(lines[2], "first") || !strings.Contains(lines[2], "0") {
		t.Fatalf("expected first task in line 2, got: %s", lines[2])
	}
	if !strings.Contains(lines[3], "second") || !strings.Contains(lines[3], "5") {
		t.Fatalf("expected second task in line 3, got: %s", lines[3])
	}
}

func TestVisualizeScheduleSortedByPriority(t *testing.T) {
	tasks := []Task{
		{ID: "a", Name: "low-pri", Priority: 10},
		{ID: "b", Name: "high-pri", Priority: 1},
	}
	result := VisualizeSchedule(tasks)
	idxHigh := strings.Index(result, "high-pri")
	idxLow := strings.Index(result, "low-pri")
	if idxHigh > idxLow {
		t.Fatal("expected high-pri before low-pri in visualization")
	}
}

// --- Additional edge case tests ---

func TestDAGSchedulerMultipleRoots(t *testing.T) {
	ds := NewDAGScheduler()
	input := []Task{
		{ID: "a", Name: "a"},
		{ID: "b", Name: "b"},
		{ID: "c", Name: "c"},
	}
	result := ds.scheduleFromTasks(input)
	if len(result) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(result))
	}
}

func TestCriticalPathZeroDuration(t *testing.T) {
	tasks := []Task{
		{ID: "a", Duration: 0},
		{ID: "b", Dependencies: []string{"a"}, Duration: 0},
	}
	path, err := CriticalPath(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) < 1 {
		t.Fatal("expected at least 1 task in path")
	}
}

func TestEstimateDurationWithInvalidDeps(t *testing.T) {
	tasks := []Task{
		{ID: "a", Dependencies: []string{"missing"}},
	}
	d := EstimateDuration(tasks)
	if d != 0 {
		t.Fatalf("expected 0 for invalid graph, got %v", d)
	}
}

func TestResourceAwareStrategySingleLargeWeight(t *testing.T) {
	ra := ResourceAwareStrategy{Capacity: 3}
	tasks := []Task{
		{ID: "a", Weight: 10},
	}
	groups := ra.Apply(tasks)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group even if over capacity, got %d", len(groups))
	}
}
