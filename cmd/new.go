package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/jackreid/task/internal/id"
	"github.com/jackreid/task/internal/model"
)

// labelList is a custom flag type for collecting multiple labels
type labelList []string

func (l *labelList) String() string {
	return strings.Join(*l, ", ")
}

func (l *labelList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

func runNew(args []string) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var description string
	var labels labelList
	var taskType string

	fs.StringVar(&description, "d", "", "Task description")
	fs.StringVar(&description, "description", "", "Task description")
	fs.Var(&labels, "l", "Label to add (can be specified multiple times)")
	fs.Var(&labels, "label", "Label to add (can be specified multiple times)")
	fs.StringVar(&taskType, "t", "task", "Task type (task, bug, feature)")
	fs.StringVar(&taskType, "type", "task", "Task type (task, bug, feature)")

	fs.Usage = func() {
		fmt.Fprintln(stderr, `Create a new task.

Usage:
  task new <title> [flags]

Flags:
  -d, --description string   Task description
  -l, --label string         Label to add (can be specified multiple times)
  -t, --type string          Task type: task, bug, feature (default "task")

Examples:
  task new "Implement login"
  task new "Fix bug" -t bug -l urgent
  task new "Add feature" -t feature -d "Detailed description" -l frontend -l priority`)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		errorf("Error: task title is required")
		fs.Usage()
		return fmt.Errorf("task title is required")
	}

	title := fs.Arg(0)

	// Validate task type
	tt, err := model.ParseTaskType(taskType)
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	s := getStore()

	// Get existing IDs to ensure uniqueness
	existingIDs, err := s.GetExistingIDs()
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	// Generate unique ID
	taskID, err := id.GenerateUnique(existingIDs)
	if err != nil {
		errorf("Error generating ID: %v", err)
		return err
	}

	// Create the task
	task := model.NewTask(taskID, title, tt)

	// Set optional fields
	if description != "" {
		task.SetDescription(description)
	}

	for _, label := range labels {
		task.AddLabel(label)
	}

	// Save the task
	if err := s.Add(task); err != nil {
		errorf("Error: %v", err)
		return err
	}

	fmt.Fprintf(stdout, "Created task %s: %s\n", task.ID, task.Title)
	return nil
}
