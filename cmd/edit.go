package cmd

import (
	"flag"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackreid/task/internal/model"
)

func runEdit(args []string) error {
	fs := flag.NewFlagSet("edit", flag.ContinueOnError)
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
		fmt.Fprintln(stderr, `Edit a task in $EDITOR or update directly with flags.

Usage:
  task edit <id> [flags]
  task edit <id>           (opens $EDITOR when no flags are provided)

Flags:
  -n, --name string        New task name
  -d, --description string Task description
  -l, --label string       Label to add (can be specified multiple times)
  -t, --type string        Task type: task, bug, feature
  -s, --status string      Task status: todo, progress, blocked, abandon, done

Examples:
  task edit abc
  task edit abc --type bug
  task edit abc -n "New name" -s done`)
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

	// If flags are provided, update directly without editor
	if fs.NFlag() > 0 {
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

	// No flags provided, launch editor
	desc := ""
	if task.Description != nil {
		desc = *task.Description
	}

	template := renderTaskTemplate(task.Title, task.Type, task.Status, task.Labels, desc)
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

	title := task.Title
	if fm.HasTitle {
		title = fm.Title
	}
	if strings.TrimSpace(title) == "" {
		errorf("Error: task title is required")
		return fmt.Errorf("task title is required")
	}

	tt := task.Type.String()
	if fm.HasType {
		tt = fm.Type
	}
	parsedType, err := model.ParseTaskType(tt)
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	statusValue := task.Status.String()
	if fm.HasStatus {
		statusValue = fm.Status
	}
	parsedStatus, err := model.ParseStatus(statusValue)
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	parsedLabels := task.Labels
	if fm.HasLabels {
		parsedLabels = normalizeLabels(fm.Labels)
	}

	descriptionValue := normalizeDescription(body)

	if title != task.Title {
		task.SetTitle(title)
	}
	if parsedType != task.Type {
		task.SetType(parsedType)
	}
	if parsedStatus != task.Status {
		task.SetStatus(parsedStatus)
	}
	if !reflect.DeepEqual(parsedLabels, task.Labels) {
		task.SetLabels(parsedLabels)
	}
	if !descriptionsEqual(descriptionValue, task.Description) {
		task.SetDescriptionValue(descriptionValue)
	}

	if err := s.Update(task); err != nil {
		errorf("Error: %v", err)
		return err
	}

	fmt.Fprintf(stdout, "Updated task %s\n", task.ID)
	return nil
}

func descriptionsEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
