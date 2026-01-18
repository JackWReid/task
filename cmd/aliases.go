package cmd

import (
	"fmt"

	"github.com/jackreid/task/internal/model"
)

// runReady lists tasks with status 'todo'
// Alias for: task list -s todo
func runReady(args []string) error {
	return runList(append([]string{"-s", "todo"}, args...))
}

// runTake sets a task status to 'progress'
// Alias for: task update $id -s progress
func runTake(args []string) error {
	if len(args) < 1 {
		errorf("Error: task ID is required")
		fmt.Fprintln(stderr, "Usage: task take <id>")
		return fmt.Errorf("task ID is required")
	}
	return updateTaskStatus(args[0], model.StatusProgress)
}

// runComplete sets a task status to 'done'
// Alias for: task update $id -s done
func runComplete(args []string) error {
	if len(args) < 1 {
		errorf("Error: task ID is required")
		fmt.Fprintln(stderr, "Usage: task complete <id>")
		return fmt.Errorf("task ID is required")
	}
	return updateTaskStatus(args[0], model.StatusDone)
}

// runBlock sets a task status to 'blocked'
// Alias for: task update $id -s blocked
func runBlock(args []string) error {
	if len(args) < 1 {
		errorf("Error: task ID is required")
		fmt.Fprintln(stderr, "Usage: task block <id>")
		return fmt.Errorf("task ID is required")
	}
	return updateTaskStatus(args[0], model.StatusBlocked)
}

// runAbandon sets a task status to 'abandon'
// Alias for: task update $id -s abandon
func runAbandon(args []string) error {
	if len(args) < 1 {
		errorf("Error: task ID is required")
		fmt.Fprintln(stderr, "Usage: task abandon <id>")
		return fmt.Errorf("task ID is required")
	}
	return updateTaskStatus(args[0], model.StatusAbandon)
}

// updateTaskStatus is a helper that updates a task's status
func updateTaskStatus(taskID string, status model.Status) error {
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

	if err := task.SetStatus(status); err != nil {
		errorf("Error: %v", err)
		return err
	}

	if err := s.Update(task); err != nil {
		errorf("Error: %v", err)
		return err
	}

	fmt.Fprintf(stdout, "Updated task %s to %s\n", task.ID, status)
	return nil
}
