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
		if term.IsTerminal(int(os.Stdout.Fd())) {
			runTUI(c)
		} else {
			c.List()
		}
		return
	}

	switch args[0] {
	case "list":
		runClient("list")
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
	case "list":
		c.List()
	case "push":
		c.Push(args[0])
	case "pop":
		c.Pop()
	case "switch":
		c.Switch()
	case "queue":
		c.Queue(args[0])
	}
}

func printUsage() {
	fmt.Println(`memo - task stack manager

Usage:
  memo                    Interactive task reorder (or show stack if non-interactive)
  memo list               Show current task stack
  memo push <description> Push a new task onto the stack
  memo pop                Pop the current task off the stack
  memo switch             Swap the top two tasks
  memo queue <description> Add a task to the bottom of the stack
  memo --help             Show this help message`)
}
