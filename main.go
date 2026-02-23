package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

const Version = "0.2.1"

func connectClient() *memoClient {
	ensureDaemon()
	c := newClient()

	resp, err := c.http.Get("http://memo/version")
	needRestart := err != nil || resp.StatusCode != 200
	if !needRestart {
		var v struct {
			Version string `json:"version"`
		}
		if decErr := json.NewDecoder(resp.Body).Decode(&v); decErr != nil || v.Version != Version {
			needRestart = true
		}
		resp.Body.Close()
	}

	if needRestart {
		killDaemon()
		ensureDaemon()
		c = newClient()
	}
	return c
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		c := connectClient()
		c.Current()
		return
	}

	switch args[0] {
	case "stack":
		c := connectClient()
		if term.IsTerminal(int(os.Stdout.Fd())) {
			runTUI(c)
		} else {
			c.Stack()
		}
		return
	case "push":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: memo push <description>")
			os.Exit(1)
		}
		description := strings.Join(args[1:], " ")
		runClient("push", description)
	case "pop":
		runClient("pop")
	case "drop":
		runClient("drop")
	case "switch":
		runClient("switch")
	case "log":
		runClient("log")
	case "history":
		runClient("history")
	case "queue":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: memo queue <description>")
			os.Exit(1)
		}
		description := strings.Join(args[1:], " ")
		runClient("queue", description)
	case "__daemon":
		runDaemon()
	case "--help", "-h", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func runClient(command string, args ...string) {
	c := connectClient()
	switch command {
	case "push":
		c.Push(args[0])
	case "pop":
		c.Pop()
	case "drop":
		c.Drop()
	case "switch":
		c.Switch()
	case "queue":
		c.Queue(args[0])
	case "log":
		c.Log()
	case "history":
		c.History()
	}
}

func printUsage() {
	fmt.Println(`memo - task stack manager

Usage:
  memo                    Show current task
  memo stack              Interactive task reorder (or show stack if non-interactive)
  memo push <description> Push a new task onto the stack
  memo pop                Pop the current task off the stack
  memo drop               Drop the current task without completing it
  memo switch             Swap the top two tasks
  memo queue <description> Add a task to the bottom of the stack
  memo log                Show all task activity log
  memo history            Show completed tasks with durations
  memo --help             Show this help message`)
}
