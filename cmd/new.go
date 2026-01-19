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
  task new [title]           (opens $EDITOR when no flags are provided)

Flags:
  -d, --description string   Task description
  -l, --label string         Label to add (can be specified multiple times)
  -t, --type string          Task type: task, bug, feature (default "task")

Examples:
  task new "Implement login"
  task new
  task new "Fix bug" -t bug -l urgent
  task new "Add feature" -t feature -d "Detailed description" -l frontend -l priority`)
	}

	// Reorder args to allow positional arguments before flags
	reorderedArgs := reorderArgsForFlexibleFlags(args)
	if err := fs.Parse(reorderedArgs); err != nil {
		return err
	}

	// If no flags are provided, check if we should launch the editor
	if fs.NFlag() == 0 {
		if fs.NArg() > 1 {
			errorf("Error: too many arguments")
			fs.Usage()
			return fmt.Errorf("too many arguments")
		}
		// Only launch editor if no arguments are provided
		if fs.NArg() == 0 {
			return runNewWithEditor("")
		}
		// If there's an argument but no flags, create the task directly
		title := fs.Arg(0)
		taskType := "task" // default type
		
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

		// Save the task
		if err := s.Add(task); err != nil {
			errorf("Error: %v", err)
			return err
		}

		fmt.Fprintf(stdout, "Created task %s: %s\n", task.ID, task.Title)
		return nil
	}

	if fs.NArg() < 1 {
		errorf("Error: task title is required")
		fs.Usage()
		return fmt.Errorf("task title is required")
	}

	if fs.NArg() > 1 {
		errorf("Error: too many arguments")
		fs.Usage()
		return fmt.Errorf("too many arguments")
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

func runNewWithEditor(title string) error {
	s := getStore()

	existingIDs, err := s.GetExistingIDs()
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	taskID, err := id.GenerateUnique(existingIDs)
	if err != nil {
		errorf("Error generating ID: %v", err)
		return err
	}

	template := renderTaskTemplate(title, model.TypeTask, model.StatusTodo, nil, "")
	edited, err := openEditorWithTemplate(template)
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	fm, body, err := parseFrontmatter(edited)
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	if fm.HasTitle {
		title = fm.Title
	}
	if strings.TrimSpace(title) == "" {
		errorf("Error: task title is required")
		return fmt.Errorf("task title is required")
	}

	taskType := model.TypeTask.String()
	if fm.HasType {
		taskType = fm.Type
	}
	tt, err := model.ParseTaskType(taskType)
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	statusValue := model.StatusTodo.String()
	if fm.HasStatus {
		statusValue = fm.Status
	}
	status, err := model.ParseStatus(statusValue)
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	labels := []string{}
	if fm.HasLabels {
		labels = normalizeLabels(fm.Labels)
	}

	task := model.NewTask(taskID, title, tt)
	if status != model.StatusTodo {
		if err := task.SetStatus(status); err != nil {
			errorf("Error: %v", err)
			return err
		}
	}
	if len(labels) > 0 {
		task.SetLabels(labels)
	}
	description := normalizeDescription(body)
	if description != nil {
		task.SetDescriptionValue(description)
	}

	if err := s.Add(task); err != nil {
		errorf("Error: %v", err)
		return err
	}

	fmt.Fprintf(stdout, "Created task %s: %s\n", task.ID, task.Title)
	return nil
}
