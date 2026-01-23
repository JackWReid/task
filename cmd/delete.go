package cmd

import (
	"flag"
	"fmt"
)

func runDelete(args []string) error {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.SetOutput(stderr)

	fs.Usage = func() {
		fmt.Fprintln(stderr, `Delete a task completely from the store.

Usage:
  task delete <id>

Examples:
  task delete abc`)
	}

	// Reorder args to allow positional arguments before flags
	reorderedArgs := reorderArgsForFlexibleFlags(args)
	if err := fs.Parse(reorderedArgs); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		errorf("Error: task ID is required")
		fs.Usage()
		return fmt.Errorf("task ID is required")
	}

	taskID := fs.Arg(0)

	s := getStore()

	task, err := s.FindByID(taskID)
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	if task == nil {
		errorf("Error: task not found: %s", taskID)
		return fmt.Errorf("task not found: %s", taskID)
	}

	if err := s.Delete(taskID); err != nil {
		errorf("Error: %v", err)
		return err
	}

	fmt.Fprintf(stdout, "Deleted task %s\n", taskID)
	return nil
}
