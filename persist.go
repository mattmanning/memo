package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func SaveState(stack *TaskStack, path string) error {
	data, err := json.MarshalIndent(stack, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func LoadState(path string) (*TaskStack, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &TaskStack{}, nil
		}
		return nil, err
	}
	var stack TaskStack
	if err := json.Unmarshal(data, &stack); err != nil {
		return nil, err
	}
	if stack.Tasks == nil {
		stack.Tasks = []Task{}
	}
	return &stack, nil
}

type LogEntry struct {
	Task    string `json:"task"`
	Started string `json:"started"`
	Stopped string `json:"stopped"`
	Reason  string `json:"reason"`
}

func LogTaskStop(path string, task Task, stoppedAt time.Time, reason string) error {
	entry := LogEntry{
		Task:    task.Description,
		Started: task.StartedAt.Format(time.RFC3339),
		Stopped: stoppedAt.Format(time.RFC3339),
		Reason:  reason,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}
