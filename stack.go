package main

import (
	"encoding/json"
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

func (s *TaskStack) MarshalJSON() ([]byte, error) {
	type Alias TaskStack
	return json.Marshal(&struct{ *Alias }{Alias: (*Alias)(s)})
}
