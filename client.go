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

func (c *memoClient) Stack() {
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

func (c *memoClient) Current() {
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

	top := stack.List()[0]
	fmt.Printf("%s (working for %s)\n", top.Description, formatDuration(time.Since(top.StartedAt)))
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

func (c *memoClient) Switch() {
	resp, err := c.http.Post("http://memo/switch", "application/json", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		fmt.Println("Need at least 2 tasks to switch.")
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "error: server returned %s\n", resp.Status)
		os.Exit(1)
	}

	var result struct {
		Started Task `json:"started"`
		Paused  Task `json:"paused"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Paused: %s\n", result.Paused.Description)
	fmt.Printf("Resuming: %s\n", result.Started.Description)
}

func (c *memoClient) Queue(description string) {
	body := strings.NewReader(fmt.Sprintf(`{"description":%q}`, description))
	resp, err := c.http.Post("http://memo/queue", "application/json", body)
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
		Queued  Task  `json:"queued"`
		Current *Task `json:"current,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Queued: %s\n", result.Queued.Description)
}

func (c *memoClient) Reorder(order []int) error {
	orderJSON, err := json.Marshal(struct {
		Order []int `json:"order"`
	}{Order: order})
	if err != nil {
		return err
	}
	resp, err := c.http.Post("http://memo/reorder", "application/json", strings.NewReader(string(orderJSON)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s", resp.Status)
	}
	return nil
}

func (c *memoClient) FetchStack() (*TaskStack, error) {
	resp, err := c.http.Get("http://memo/stack")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stack TaskStack
	if err := json.NewDecoder(resp.Body).Decode(&stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (c *memoClient) fetchLog() []LogEntry {
	resp, err := c.http.Get("http://memo/log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "error: server returned %s\n", resp.Status)
		os.Exit(1)
	}

	var entries []LogEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	return entries
}

func (c *memoClient) Log() {
	entries := c.fetchLog()
	if len(entries) == 0 {
		fmt.Println("No log entries yet.")
		return
	}
	for _, e := range entries {
		stopped, _ := time.Parse(time.RFC3339, e.Stopped)
		started, _ := time.Parse(time.RFC3339, e.Started)
		dur := stopped.Sub(started)
		fmt.Printf("[%s] %-10s \"%s\" (worked %s)\n",
			stopped.Local().Format("2006-01-02 15:04"),
			e.Reason,
			e.Task,
			formatDuration(dur))
	}
}

func (c *memoClient) History() {
	entries := c.fetchLog()
	var popped []LogEntry
	for _, e := range entries {
		if e.Reason == "popped" {
			popped = append(popped, e)
		}
	}
	if len(popped) == 0 {
		fmt.Println("No completed tasks yet.")
		return
	}
	for _, e := range popped {
		started, _ := time.Parse(time.RFC3339, e.Started)
		stopped, _ := time.Parse(time.RFC3339, e.Stopped)
		dur := stopped.Sub(started)
		fmt.Printf("%s\n  Started:  %s\n  Finished: %s\n  Duration: %s\n",
			e.Task,
			started.Local().Format("2006-01-02 15:04"),
			stopped.Local().Format("2006-01-02 15:04"),
			formatDuration(dur))
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
