package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func memoDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(home, ".memo")
}

func socketPath() string {
	return filepath.Join(memoDir(), "memo.sock")
}

func pidPath() string {
	return filepath.Join(memoDir(), "memo.pid")
}

func statePath() string {
	return filepath.Join(memoDir(), "state.json")
}

func logPath() string {
	return filepath.Join(memoDir(), "log.jsonl")
}

func runDaemon() {
	dir := memoDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	stack, err := LoadState(statePath())
	if err != nil {
		log.Fatalf("failed to load state: %v", err)
	}

	sock := socketPath()
	// Clean up stale socket
	if _, err := os.Stat(sock); err == nil {
		os.Remove(sock)
	}

	ln, err := net.Listen("unix", sock)
	if err != nil {
		log.Fatalf("failed to listen on socket: %v", err)
	}

	// Write PID file
	if err := os.WriteFile(pidPath(), []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		log.Fatalf("failed to write PID file: %v", err)
	}

	var mu sync.Mutex

	mux := http.NewServeMux()

	mux.HandleFunc("/stack", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stack)
	})

	mux.HandleFunc("/push", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Description) == "" {
			http.Error(w, "description required", http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		now := time.Now().UTC()
		var paused *Task
		if top := stack.Peek(); top != nil {
			copy := *top
			paused = &copy
			LogTaskStop(logPath(), *top, now, "pushed")
		}

		stack.Push(req.Description)
		SaveState(stack, statePath())

		resp := struct {
			Started Task  `json:"started"`
			Paused  *Task `json:"paused,omitempty"`
		}{
			Started: *stack.Peek(),
			Paused:  paused,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/pop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		popped := stack.Pop()
		if popped == nil {
			http.Error(w, "stack is empty", http.StatusBadRequest)
			return
		}

		now := time.Now().UTC()
		LogTaskStop(logPath(), *popped, now, "popped")
		SaveState(stack, statePath())

		var resuming *Task
		if top := stack.Peek(); top != nil {
			copy := *top
			resuming = &copy
		}

		resp := struct {
			Popped   Task  `json:"popped"`
			Resuming *Task `json:"resuming,omitempty"`
		}{
			Popped:   *popped,
			Resuming: resuming,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/switch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		started, paused := stack.Switch()
		if started == nil {
			http.Error(w, "need at least 2 tasks to switch", http.StatusBadRequest)
			return
		}

		now := time.Now().UTC()
		LogTaskStop(logPath(), *paused, now, "switched")
		SaveState(stack, statePath())

		resp := struct {
			Started Task `json:"started"`
			Paused  Task `json:"paused"`
		}{
			Started: *started,
			Paused:  *paused,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/queue", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Description) == "" {
			http.Error(w, "description required", http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		queued := stack.Queue(req.Description)
		SaveState(stack, statePath())

		resp := struct {
			Queued  Task  `json:"queued"`
			Current *Task `json:"current,omitempty"`
		}{
			Queued:  *queued,
			Current: stack.Peek(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/reorder", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Order []int `json:"order"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		oldTop := stack.Peek()
		var oldTopDesc string
		if oldTop != nil {
			oldTopDesc = oldTop.Description
		}

		if err := stack.Reorder(req.Order); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		now := time.Now().UTC()
		newTop := stack.Peek()
		if oldTop != nil && newTop != nil && oldTopDesc != newTop.Description {
			LogTaskStop(logPath(), Task{Description: oldTopDesc, StartedAt: oldTop.StartedAt}, now, "reordered")
		}
		SaveState(stack, statePath())

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stack)
	})

	// Handle signals for clean shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		ln.Close()
		os.Remove(sock)
		os.Remove(pidPath())
		os.Exit(0)
	}()

	server := &http.Server{Handler: mux}
	if err := server.Serve(ln); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		log.Fatalf("server error: %v", err)
	}
}

func ensureDaemon() {
	sock := socketPath()

	// Try connecting to existing daemon
	if tryConnect(sock) {
		return
	}

	// Check PID file
	if data, err := os.ReadFile(pidPath()); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			if process, err := os.FindProcess(pid); err == nil {
				if err := process.Signal(syscall.Signal(0)); err == nil {
					// Process exists, wait for socket
					for i := 0; i < 20; i++ {
						time.Sleep(100 * time.Millisecond)
						if tryConnect(sock) {
							return
						}
					}
				}
			}
		}
	}

	// Start daemon
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("failed to find executable: %v", err)
	}

	cmd := exec.Command(exe, "__daemon")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		log.Fatalf("failed to start daemon: %v", err)
	}
	cmd.Process.Release()

	// Poll for daemon readiness
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		if tryConnect(sock) {
			return
		}
	}
	fmt.Fprintf(os.Stderr, "error: daemon did not start in time\n")
	os.Exit(1)
}

func tryConnect(sock string) bool {
	conn, err := net.DialTimeout("unix", sock, 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
