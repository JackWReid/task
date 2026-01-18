package cmd

import (
	"flag"
	"fmt"

	"github.com/jackreid/task/internal/model"
)

func runUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var name string
	var description string
	var labels labelList
	var taskType string
	var status string

	fs.StringVar(&name, "n", "", "New task name")
	fs.StringVar(&name, "name", "", "New task name")
	fs.StringVar(&description, "d", "", "Task description")
	fs.StringVar(&description, "description", "", "Task description")
	fs.Var(&labels, "l", "Label to add (can be specified multiple times)")
	fs.Var(&labels, "label", "Label to add (can be specified multiple times)")
	fs.StringVar(&taskType, "t", "", "Task type (task, bug, feature)")
	fs.StringVar(&taskType, "type", "", "Task type (task, bug, feature)")
	fs.StringVar(&status, "s", "", "Task status (todo, progress, blocked, abandon, done)")
	fs.StringVar(&status, "status", "", "Task status (todo, progress, blocked, abandon, done)")

	fs.Usage = func() {
		fmt.Fprintln(stderr, `Update an existing task.

Usage:
  task update <id> [flags]

Flags:
  -n, --name string        New task name
  -d, --description string Task description
  -l, --label string       Label to add (can be specified multiple times)
  -t, --type string        Task type: task, bug, feature
  -s, --status string      Task status: todo, progress, blocked, abandon, done

Examples:
  task update abc -n "New name"
  task update abc -s done
  task update abc -l urgent -l priority`)
	}

	if err := fs.Parse(args); err != nil {
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

	// Apply updates
	if name != "" {
		task.SetTitle(name)
	}

	if description != "" {
		task.SetDescription(description)
	}

	if len(labels) > 0 {
		task.SetLabels(labels)
	}

	if taskType != "" {
		tt, err := model.ParseTaskType(taskType)
		if err != nil {
			errorf("Error: %v", err)
			return err
		}
		task.SetType(tt)
	}

	if status != "" {
		st, err := model.ParseStatus(status)
		if err != nil {
			errorf("Error: %v", err)
			return err
		}
		task.SetStatus(st)
	}

	if err := s.Update(task); err != nil {
		errorf("Error: %v", err)
		return err
	}

	fmt.Fprintf(stdout, "Updated task %s\n", task.ID)
	return nil
}
