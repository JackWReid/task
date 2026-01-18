package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/jackreid/task/internal/store"
)

var (
	// stdout and stderr for output (can be overridden in tests)
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
	// stdin for input (can be overridden in tests)
	stdin io.Reader = os.Stdin
	// workDir is the working directory for the store (can be overridden in tests)
	workDir string = ""
)

// getStore returns a store instance for the current working directory
func getStore() *store.Store {
	return store.New(workDir)
}

// Execute runs the task CLI application
func Execute() error {
	if len(os.Args) < 2 {
		printHelp()
		return nil
	}
	return run(os.Args[1:])
}

// run is the internal entry point that can be tested
func run(args []string) error {
	if len(args) == 0 {
		printHelp()
		return nil
	}

	command := args[0]

	switch command {
	case "help", "-h", "--help":
		printHelp()
		return nil
	case "init":
		return runInit(args[1:])
	case "list":
		return runList(args[1:])
	case "new":
		return runNew(args[1:])
	case "update":
		return runUpdate(args[1:])
	case "show":
		return runShow(args[1:])
	case "note":
		return runNote(args[1:])
	case "ready":
		return runReady(args[1:])
	case "take":
		return runTake(args[1:])
	case "complete":
		return runComplete(args[1:])
	case "block":
		return runBlock(args[1:])
	case "abandon":
		return runAbandon(args[1:])
	default:
		fmt.Fprintf(stderr, "Unknown command: %s\n", command)
		printHelp()
		return fmt.Errorf("unknown command: %s", command)
	}
}

func printHelp() {
	fmt.Fprintln(stdout, `task - a simple task management app

Usage:
  task <command> [arguments]

Commands:
  init        Initialize task management in the current directory
  list        List all tasks
  new         Create a new task
  update      Update an existing task
  show        Show task details
  note        Add a note to a task

Aliases:
  ready       List tasks with status 'todo'
  take        Set task status to 'progress'
  complete    Set task status to 'done'
  block       Set task status to 'blocked'
  abandon     Set task status to 'abandon'

Use "task <command> -h" for more information about a command.`)
}

// errorf prints an error message to stderr
func errorf(format string, args ...interface{}) {
	fmt.Fprintf(stderr, format+"\n", args...)
}
