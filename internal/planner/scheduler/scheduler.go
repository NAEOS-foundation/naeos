package scheduler

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
)

type Scheduler interface {
	Schedule(neir any) ([]Task, error)
	ParallelGroups(tasks []Task) []ParallelGroup
}

type Task struct {
	ID           string
	Name         string
	Dependencies []string
	Priority     int
	Duration     time.Duration
	Weight       int
}

type ParallelGroup struct {
	Tasks []Task
	Level int
}

type DefaultScheduler struct{}

func NewScheduler() Scheduler {
	return DefaultScheduler{}
}

func (DefaultScheduler) Schedule(neir any) ([]Task, error) {
	if neir == nil {
		return nil, fmt.Errorf("neir is nil")
	}

	neirModel, ok := neir.(*model.NEIR)
	if !ok {
		return []Task{{ID: "task-1", Name: "bootstrap"}}, nil
	}

	var tasks []Task
	taskID := 0

	tasks = append(tasks, Task{
		ID:       fmt.Sprintf("task-%d", taskID),
		Name:     "validate-specification",
		Priority: 0,
	})
	taskID++

	tasks = append(tasks, Task{
		ID:           fmt.Sprintf("task-%d", taskID),
		Name:         "build-model",
		Dependencies: []string{fmt.Sprintf("task-%d", taskID-1)},
		Priority:     1,
	})
	taskID++

	for _, mod := range neirModel.Modules {
		dep := fmt.Sprintf("task-%d", taskID-1)
		tasks = append(tasks, Task{
			ID:           fmt.Sprintf("task-%d", taskID),
			Name:         fmt.Sprintf("generate-module-%s", mod.Name),
			Dependencies: []string{dep},
			Priority:     2,
		})
		taskID++
	}

	if len(neirModel.Modules) > 0 {
		lastModuleTask := fmt.Sprintf("task-%d", taskID-1)
		for _, svc := range neirModel.Services {
			tasks = append(tasks, Task{
				ID:           fmt.Sprintf("task-%d", taskID),
				Name:         fmt.Sprintf("generate-service-%s", svc.Name),
				Dependencies: []string{lastModuleTask},
				Priority:     3,
			})
			taskID++
		}
	}

	if len(neirModel.Services) > 0 {
		lastServiceTask := fmt.Sprintf("task-%d", taskID-1)
		tasks = append(tasks, Task{
			ID:           fmt.Sprintf("task-%d", taskID),
			Name:         "generate-config",
			Dependencies: []string{lastServiceTask},
			Priority:     4,
		})
		taskID++
	} else if len(neirModel.Modules) > 0 {
		lastModuleTask := fmt.Sprintf("task-%d", taskID-1)
		tasks = append(tasks, Task{
			ID:           fmt.Sprintf("task-%d", taskID),
			Name:         "generate-config",
			Dependencies: []string{lastModuleTask},
			Priority:     4,
		})
		taskID++
	}

	tasks = append(tasks, Task{
		ID:       fmt.Sprintf("task-%d", taskID),
		Name:     "validate-output",
		Priority: 5,
	})

	return tasks, nil
}

func (DefaultScheduler) ParallelGroups(tasks []Task) []ParallelGroup {
	if len(tasks) == 0 {
		return nil
	}

	priorityMap := make(map[int][]Task)
	for _, t := range tasks {
		priorityMap[t.Priority] = append(priorityMap[t.Priority], t)
	}

	var groups []ParallelGroup
	for level := 0; level <= 5; level++ {
		if groupTasks, ok := priorityMap[level]; ok && len(groupTasks) > 0 {
			groups = append(groups, ParallelGroup{
				Tasks: groupTasks,
				Level: level,
			})
		}
	}

	return groups
}

type DAGScheduler struct{}

func NewDAGScheduler() *DAGScheduler {
	return &DAGScheduler{}
}

type dagNode struct {
	task       Task
	inDegree   int
	children   []string
}

func (d *DAGScheduler) Schedule(neir any) ([]Task, error) {
	if neir == nil {
		return nil, fmt.Errorf("neir is nil")
	}

	neirModel, ok := neir.(*model.NEIR)
	if !ok {
		return []Task{{ID: "task-1", Name: "bootstrap"}}, nil
	}

	defaultSched := DefaultScheduler{}
	tasks, err := defaultSched.Schedule(neirModel)
	if err != nil {
		return nil, err
	}

	return d.scheduleFromTasks(tasks), nil
}

func (d *DAGScheduler) scheduleFromTasks(tasks []Task) []Task {
	nodes := buildDAG(nodesFromTasks(tasks))
	levels := topologicalLevels(nodes)

	var result []Task
	for _, level := range levels {
		sort.Slice(level, func(i, j int) bool {
			return level[i].task.Priority < level[j].task.Priority
		})
		for _, n := range level {
			result = append(result, n.task)
		}
	}
	return result
}

func (d *DAGScheduler) ParallelGroups(tasks []Task) []ParallelGroup {
	nodes := buildDAG(nodesFromTasks(tasks))
	levels := topologicalLevels(nodes)

	var groups []ParallelGroup
	for i, level := range levels {
		var levelTasks []Task
		for _, n := range level {
			levelTasks = append(levelTasks, n.task)
		}
		groups = append(groups, ParallelGroup{
			Tasks: levelTasks,
			Level: i,
		})
	}
	return groups
}

func nodesFromTasks(tasks []Task) map[string]*dagNode {
	nodes := make(map[string]*dagNode)
	for _, t := range tasks {
		nodes[t.ID] = &dagNode{task: t}
	}
	for _, t := range tasks {
		for _, dep := range t.Dependencies {
			if parent, ok := nodes[dep]; ok {
				parent.children = append(parent.children, t.ID)
			}
		}
	}
	return nodes
}

func buildDAG(nodes map[string]*dagNode) map[string]*dagNode {
	for _, node := range nodes {
		node.inDegree = 0
	}
	for _, node := range nodes {
		for _, childID := range node.children {
			if child, ok := nodes[childID]; ok {
				child.inDegree++
			}
		}
	}
	return nodes
}

func topologicalLevels(nodes map[string]*dagNode) [][]*dagNode {
	remaining := make(map[string]*dagNode)
	for k, v := range nodes {
		remaining[k] = v
	}

	var levels [][]*dagNode
	for len(remaining) > 0 {
		var currentLevel []*dagNode
		for id, node := range remaining {
			if node.inDegree == 0 {
				currentLevel = append(currentLevel, node)
				delete(remaining, id)
			}
		}
		if len(currentLevel) == 0 {
			break
		}
		levels = append(levels, currentLevel)
		for _, node := range currentLevel {
			for _, childID := range node.children {
				if child, ok := remaining[childID]; ok {
					child.inDegree--
				}
			}
		}
	}
	return levels
}

func CriticalPath(tasks []Task) ([]Task, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	if err := ValidateSchedule(tasks); err != nil {
		return nil, err
	}

	dist := make(map[string]time.Duration)
	predecessor := make(map[string]string)
	for _, t := range tasks {
		dist[t.ID] = t.Duration
	}

	ordered := topologicalOrder(tasks)
	for _, t := range ordered {
		for _, t2 := range tasks {
			for _, dep := range t2.Dependencies {
				if dep == t.ID {
					candidate := dist[t.ID] + t2.Duration
					if candidate > dist[t2.ID] {
						dist[t2.ID] = candidate
						predecessor[t2.ID] = t.ID
					}
				}
			}
		}
	}

	var maxDur time.Duration
	var endTask string
	for _, t := range tasks {
		if endTask == "" || dist[t.ID] > maxDur {
			maxDur = dist[t.ID]
			endTask = t.ID
		}
	}

	idSet := make(map[string]Task)
	for _, t := range tasks {
		idSet[t.ID] = t
	}

	var path []Task
	for endTask != "" {
		path = append([]Task{idSet[endTask]}, path...)
		endTask = predecessor[endTask]
	}
	return path, nil
}

func EstimateDuration(tasks []Task) time.Duration {
	if len(tasks) == 0 {
		return 0
	}

	path, err := CriticalPath(tasks)
	if err != nil || len(path) == 0 {
		return 0
	}

	var total time.Duration
	for _, t := range path {
		total += t.Duration
	}
	return total
}

func topologicalOrder(tasks []Task) []Task {
	inDeg := make(map[string]int)
	adj := make(map[string][]string)
	idSet := make(map[string]Task)
	for _, t := range tasks {
		idSet[t.ID] = t
		inDeg[t.ID] = len(t.Dependencies)
		for _, dep := range t.Dependencies {
			adj[dep] = append(adj[dep], t.ID)
		}
	}

	var queue []string
	for id, deg := range inDeg {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	var result []Task
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		result = append(result, idSet[id])
		var next []string
		for _, child := range adj[id] {
			inDeg[child]--
			if inDeg[child] == 0 {
				next = append(next, child)
			}
		}
		sort.Strings(next)
		queue = append(queue, next...)
	}
	return result
}

type ScheduleStrategy interface {
	Apply(tasks []Task) [][]Task
}

type PriorityStrategy struct{}

func (p PriorityStrategy) Apply(tasks []Task) [][]Task {
	if len(tasks) == 0 {
		return nil
	}

	sorted := make([]Task, len(tasks))
	copy(sorted, tasks)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})

	var groups [][]Task
	var currentGroup []Task
	currentPriority := sorted[0].Priority
	for _, t := range sorted {
		if t.Priority != currentPriority {
			groups = append(groups, currentGroup)
			currentGroup = nil
			currentPriority = t.Priority
		}
		currentGroup = append(currentGroup, t)
	}
	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}
	return groups
}

type RoundRobinStrategy struct{}

func (r RoundRobinStrategy) Apply(tasks []Task) [][]Task {
	if len(tasks) == 0 {
		return nil
	}

	sorted := make([]Task, len(tasks))
	copy(sorted, tasks)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})

	var groups [][]Task
	for i := 0; i < len(sorted); i++ {
		groups = append(groups, []Task{sorted[i]})
	}
	return groups
}

type ResourceAwareStrategy struct {
	Capacity int
}

func (ra ResourceAwareStrategy) Apply(tasks []Task) [][]Task {
	if len(tasks) == 0 || ra.Capacity <= 0 {
		return nil
	}

	sorted := make([]Task, len(tasks))
	copy(sorted, tasks)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority < sorted[j].Priority
		}
		return sorted[i].ID < sorted[j].ID
	})

	var groups [][]Task
	var currentGroup []Task
	currentWeight := 0

	for _, t := range sorted {
		if t.Weight == 0 {
			if len(currentGroup) > 0 {
				groups = append(groups, currentGroup)
				currentGroup = nil
				currentWeight = 0
			}
			groups = append(groups, []Task{t})
			continue
		}

		w := t.Weight
		if w < 0 {
			w = 0
		}
		if currentWeight+w > ra.Capacity && len(currentGroup) > 0 {
			groups = append(groups, currentGroup)
			currentGroup = nil
			currentWeight = 0
		}
		currentGroup = append(currentGroup, t)
		currentWeight += w
	}
	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}
	return groups
}

func ValidateSchedule(tasks []Task) error {
	idSet := make(map[string]bool)
	for _, t := range tasks {
		if idSet[t.ID] {
			return fmt.Errorf("duplicate task ID: %s", t.ID)
		}
		idSet[t.ID] = true
	}

	for _, t := range tasks {
		for _, dep := range t.Dependencies {
			if !idSet[dep] {
				return fmt.Errorf("task %s depends on unknown task %s", t.ID, dep)
			}
		}
	}

	visited := make(map[string]int)
	const (
		white = 0
		gray  = 1
		black = 2
	)
	for id := range idSet {
		visited[id] = white
	}

	adj := make(map[string][]string)
	for _, t := range tasks {
		for _, dep := range t.Dependencies {
			adj[dep] = append(adj[dep], t.ID)
		}
	}

	for id := range idSet {
		if visited[id] != white {
			continue
		}
		if err := dfsCycle(id, adj, visited); err != nil {
			return err
		}
	}

	return nil
}

func dfsCycle(id string, adj map[string][]string, visited map[string]int) error {
	const (
		white = 0
		gray  = 1
		black = 2
	)

	visited[id] = gray
	for _, child := range adj[id] {
		if visited[child] == gray {
			return fmt.Errorf("cycle detected involving tasks %s and %s", id, child)
		}
		if visited[child] == white {
			if err := dfsCycle(child, adj, visited); err != nil {
				return err
			}
		}
	}
	visited[id] = black
	return nil
}

type TaskGroup struct {
	Name  string
	Tasks []Task
}

func NewTaskGroup(name string, tasks []Task) TaskGroup {
	return TaskGroup{
		Name:  name,
		Tasks: tasks,
	}
}

func (tg TaskGroup) TaskIDs() []string {
	var ids []string
	for _, t := range tg.Tasks {
		ids = append(ids, t.ID)
	}
	return ids
}

func (tg TaskGroup) TotalDuration() time.Duration {
	var total time.Duration
	for _, t := range tg.Tasks {
		total += t.Duration
	}
	return total
}

func (tg TaskGroup) TotalWeight() int {
	total := 0
	for _, t := range tg.Tasks {
		total += t.Weight
	}
	return total
}

func VisualizeSchedule(tasks []Task) string {
	if len(tasks) == 0 {
		return ""
	}

	sorted := make([]Task, len(tasks))
	copy(sorted, tasks)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority < sorted[j].Priority
		}
		return sorted[i].ID < sorted[j].ID
	})

	maxNameLen := 0
	for _, t := range sorted {
		if len(t.Name) > maxNameLen {
			maxNameLen = len(t.Name)
		}
	}
	if maxNameLen < 8 {
		maxNameLen = 8
	}

	var sb strings.Builder
	header := fmt.Sprintf("%-*s  %-6s  %-8s  %s", maxNameLen, "TASK", "PRI", "DUR", "TIMELINE")
	sb.WriteString(header)
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("-", len(header)))
	sb.WriteString("\n")

	var accumulated time.Duration
	for _, t := range sorted {
		name := t.Name
		if len(name) > maxNameLen {
			name = name[:maxNameLen]
		}
		dur := t.Duration
		if dur == 0 {
			dur = 1 * time.Second
		}

		barLen := int(dur.Seconds())
		if barLen < 1 {
			barLen = 1
		}
		if barLen > 20 {
			barLen = 20
		}
		start := int(accumulated.Seconds())
		padding := strings.Repeat(" ", start)
		bar := strings.Repeat("#", barLen)

		sb.WriteString(fmt.Sprintf("%-*s  %-6d  %-8s  %s%s [%s]\n",
			maxNameLen, name, t.Priority, dur.Round(time.Millisecond), padding, bar, dur.Round(time.Millisecond)))

		accumulated += dur
	}

	return sb.String()
}
