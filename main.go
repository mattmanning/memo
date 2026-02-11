package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		runClient("list")
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
	}
}

func printUsage() {
	fmt.Println(`memo - task stack manager

Usage:
  memo                    Show current task stack
  memo push <description> Push a new task onto the stack
  memo pop                Pop the current task off the stack
  memo --help             Show this help message`)
}
