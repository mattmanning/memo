package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Task struct {
	Description string    `json:"description"`
	StartedAt   time.Time `json:"started_at"`
}

type TaskStack struct {
	Tasks []Task `json:"tasks"`
}

func (s *TaskStack) Push(description string) *Task {
	t := Task{
		Description: description,
		StartedAt:   time.Now().UTC(),
	}
	s.Tasks = append([]Task{t}, s.Tasks...)
	return &s.Tasks[0]
}

func (s *TaskStack) Pop() *Task {
	if len(s.Tasks) == 0 {
		return nil
	}
	t := s.Tasks[0]
	s.Tasks = s.Tasks[1:]
	return &t
}

func (s *TaskStack) Peek() *Task {
	if len(s.Tasks) == 0 {
		return nil
	}
	return &s.Tasks[0]
}

func (s *TaskStack) List() []Task {
	return s.Tasks
}

func (s *TaskStack) Len() int {
	return len(s.Tasks)
}

func (s *TaskStack) Switch() (started, paused *Task) {
	if len(s.Tasks) < 2 {
		return nil, nil
	}
	s.Tasks[0], s.Tasks[1] = s.Tasks[1], s.Tasks[0]
	return &s.Tasks[0], &s.Tasks[1]
}

func (s *TaskStack) Queue(description string) *Task {
	t := Task{
		Description: description,
		StartedAt:   time.Now().UTC(),
	}
	s.Tasks = append(s.Tasks, t)
	return &s.Tasks[len(s.Tasks)-1]
}

func (s *TaskStack) Reorder(order []int) error {
	if len(order) != len(s.Tasks) {
		return fmt.Errorf("order length %d does not match stack length %d", len(order), len(s.Tasks))
	}
	seen := make(map[int]bool, len(order))
	for _, idx := range order {
		if idx < 0 || idx >= len(s.Tasks) {
			return fmt.Errorf("index %d out of range", idx)
		}
		if seen[idx] {
			return fmt.Errorf("duplicate index %d", idx)
		}
		seen[idx] = true
	}
	reordered := make([]Task, len(s.Tasks))
	for i, idx := range order {
		reordered[i] = s.Tasks[idx]
	}
	s.Tasks = reordered
	return nil
}

func (s *TaskStack) MarshalJSON() ([]byte, error) {
	type Alias TaskStack
	return json.Marshal(&struct{ *Alias }{Alias: (*Alias)(s)})
}
