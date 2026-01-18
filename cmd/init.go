package cmd

import (
	"flag"
	"fmt"
)

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, `Initialize task management in the current directory.

Usage:
  task init

This creates a .task/ directory and an empty task.json file.`)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	s := getStore()
	if err := s.Init(); err != nil {
		errorf("Error: %v", err)
		return err
	}

	fmt.Fprintln(stdout, "Initialized task management in .task/")
	return nil
}
