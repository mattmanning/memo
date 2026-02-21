package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		ensureDaemon()
		c := newClient()
		c.Current()
		return
	}

	switch args[0] {
	case "stack":
		ensureDaemon()
		c := newClient()
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
	ensureDaemon()
	c := newClient()
	switch command {
	case "push":
		c.Push(args[0])
	case "pop":
		c.Pop()
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
  memo switch             Swap the top two tasks
  memo queue <description> Add a task to the bottom of the stack
  memo log                Show all task activity log
  memo history            Show completed tasks with durations
  memo --help             Show this help message`)
}
