package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/jackreid/task/internal/store"
	"github.com/jackreid/task/internal/version"
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
	case "version", "-v", "--version":
		return runVersion(args[1:])
	case "init":
		return runInit(args[1:])
	case "list":
		return runList(args[1:])
	case "new":
		return runNew(args[1:])
	case "edit":
		return runEdit(args[1:])
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
  edit        Edit a task in $EDITOR
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

// runVersion prints version information
func runVersion(args []string) error {
	fmt.Fprintln(stdout, version.String())
	return nil
}

// errorf prints an error message to stderr
func errorf(format string, args ...interface{}) {
	fmt.Fprintf(stderr, format+"\n", args...)
}

// reorderArgsForFlexibleFlags moves the first positional argument (non-flag) to the end
// of the args slice, allowing flags to come after positional arguments.
// This enables commands like: task new "Title" --type bug
func reorderArgsForFlexibleFlags(args []string) []string {
	if len(args) == 0 {
		return args
	}

	// Find the first non-flag argument
	posArgIndex := -1
	for i := 0; i < len(args); i++ {
		arg := args[i]
		// Check if this is a flag (starts with -)
		if len(arg) > 0 && arg[0] == '-' {
			// This is a flag
			// Check if it's a short flag with value attached (e.g., -tbug) or long flag
			// For flags like -t or --type, the next arg might be the value
			// We'll assume the next non-flag arg is the value if it exists
			if i+1 < len(args) && len(args[i+1]) > 0 && args[i+1][0] != '-' {
				// Skip the flag value
				i++
			}
			continue
		}
		// Found first positional argument
		posArgIndex = i
		break
	}

	// If no positional argument found, return as-is
	if posArgIndex == -1 {
		return args
	}

	// Reorder: flags before positional, then positional, then flags after
	posArg := args[posArgIndex]
	flagsBefore := args[:posArgIndex]
	flagsAfter := args[posArgIndex+1:]

	// Combine: flags before, flags after, then positional argument
	result := make([]string, 0, len(args))
	result = append(result, flagsBefore...)
	result = append(result, flagsAfter...)
	result = append(result, posArg)

	return result
}
