package cmd

import (
	"flag"
	"fmt"
)

func runClean(args []string) error {
	fs := flag.NewFlagSet("clean", flag.ContinueOnError)
	fs.SetOutput(stderr)

	fs.Usage = func() {
		fmt.Fprintln(stderr, `Delete all closed tasks from the store.

Closed tasks are those with status 'done' or 'abandon'.

Usage:
  task clean

Examples:
  task clean`)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	s := getStore()

	deleted, err := s.Clean()
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	if deleted == 0 {
		fmt.Fprintln(stdout, "No closed tasks to delete")
	} else {
		fmt.Fprintf(stdout, "Deleted %d closed task(s)\n", deleted)
	}

	return nil
}
