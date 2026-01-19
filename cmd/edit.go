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
	fs.Usage = func() {
		fmt.Fprintln(stderr, `Edit a task in $EDITOR.

Usage:
  task edit <id>

Examples:
  task edit abc`)
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

	description := ""
	if task.Description != nil {
		description = *task.Description
	}

	template := renderTaskTemplate(task.Title, task.Type, task.Status, task.Labels, description)
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

	taskType := task.Type.String()
	if fm.HasType {
		taskType = fm.Type
	}
	tt, err := model.ParseTaskType(taskType)
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	statusValue := task.Status.String()
	if fm.HasStatus {
		statusValue = fm.Status
	}
	status, err := model.ParseStatus(statusValue)
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	labels := task.Labels
	if fm.HasLabels {
		labels = normalizeLabels(fm.Labels)
	}

	descriptionValue := normalizeDescription(body)

	if title != task.Title {
		task.SetTitle(title)
	}
	if tt != task.Type {
		task.SetType(tt)
	}
	if status != task.Status {
		task.SetStatus(status)
	}
	if !reflect.DeepEqual(labels, task.Labels) {
		task.SetLabels(labels)
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
