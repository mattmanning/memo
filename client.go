package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type memoClient struct {
	http *http.Client
}

func newClient() *memoClient {
	sock := socketPath()
	return &memoClient{
		http: &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", sock)
				},
			},
		},
	}
}

func (c *memoClient) List() {
	resp, err := c.http.Get("http://memo/stack")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var stack TaskStack
	if err := json.NewDecoder(resp.Body).Decode(&stack); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if stack.Len() == 0 {
		fmt.Println("No tasks. Use \"memo push <description>\" to start one.")
		return
	}

	for i, task := range stack.List() {
		if i == 0 {
			fmt.Printf("\u2192 %s (working for %s)\n", task.Description, formatDuration(time.Since(task.StartedAt)))
		} else {
			fmt.Printf("  %s (paused)\n", task.Description)
		}
	}
}

func (c *memoClient) Push(description string) {
	body := strings.NewReader(fmt.Sprintf(`{"description":%q}`, description))
	resp, err := c.http.Post("http://memo/push", "application/json", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "error: server returned %s\n", resp.Status)
		os.Exit(1)
	}

	var result struct {
		Started Task  `json:"started"`
		Paused  *Task `json:"paused,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if result.Paused != nil {
		fmt.Printf("Paused: %s\n", result.Paused.Description)
	}
	fmt.Printf("Started: %s\n", result.Started.Description)
}

func (c *memoClient) Pop() {
	resp, err := c.http.Post("http://memo/pop", "application/json", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		fmt.Println("No tasks to pop.")
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "error: server returned %s\n", resp.Status)
		os.Exit(1)
	}

	var result struct {
		Popped   Task  `json:"popped"`
		Resuming *Task `json:"resuming,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	duration := time.Since(result.Popped.StartedAt)
	fmt.Printf("Done: %s (%s)\n", result.Popped.Description, formatDuration(duration))

	if result.Resuming != nil {
		fmt.Printf("Resuming: %s\n", result.Resuming.Description)
	} else {
		fmt.Println("No more tasks.")
	}
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}
