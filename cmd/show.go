package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/jackreid/task/internal/model"
)

func runShow(args []string) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var jsonOutput bool

	fs.BoolVar(&jsonOutput, "json", false, "Output as JSON")

	fs.Usage = func() {
		fmt.Fprintln(stderr, `Show task details.

Usage:
  task show <id> [flags]

Flags:
  --json  Output as JSON

Examples:
  task show abc
  task show abc --json`)
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

	if jsonOutput {
		return printTaskJSON(task)
	}

	return printTaskDetail(task)
}

func printTaskJSON(task *model.Task) error {
	// For JSON output, we bypass the custom MarshalJSON to get clean output
	type TaskJSON struct {
		ID          string       `json:"id"`
		CreatedAt   string       `json:"created_at"`
		UpdatedAt   string       `json:"updated_at"`
		Title       string       `json:"title"`
		Description *string      `json:"description"`
		Type        string       `json:"type"`
		Status      string       `json:"status"`
		Labels      []string     `json:"labels"`
		Notes       []model.Note `json:"notes"`
	}
	t := TaskJSON{
		ID:          task.ID,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Title:       task.Title,
		Description: task.Description,
		Type:        string(task.Type),
		Status:      string(task.Status),
		Labels:      task.Labels,
		Notes:       task.Notes,
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, string(data))
	return nil
}

func printTaskDetail(task *model.Task) error {
	statusColor := getStatusColor(task.Status)
	typeIcon := getTypeIcon(task.Type)

	// Header
	fmt.Fprintf(stdout, "%s%s%s %s\n", colorCyan, task.ID, colorReset, task.Title)
	fmt.Fprintln(stdout, strings.Repeat("─", 40))

	// Status and Type
	fmt.Fprintf(stdout, "Status:  %s%s %s%s\n", statusColor, statusSymbol(task.Status), task.Status, colorReset)
	fmt.Fprintf(stdout, "Type:    %s %s\n", typeIcon, task.Type)

	// Labels
	if len(task.Labels) > 0 {
		fmt.Fprintf(stdout, "Labels:  %s\n", strings.Join(task.Labels, ", "))
	} else {
		fmt.Fprintf(stdout, "Labels:  %s(none)%s\n", colorGray, colorReset)
	}

	// Description
	fmt.Fprintln(stdout)
	if task.Description != nil && *task.Description != "" {
		fmt.Fprintf(stdout, "Description:\n  %s\n", *task.Description)
	} else {
		fmt.Fprintf(stdout, "Description: %s(none)%s\n", colorGray, colorReset)
	}

	// Timestamps
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "%sCreated: %s%s\n", colorGray, task.CreatedAt.Format("2006-01-02 15:04:05"), colorReset)
	fmt.Fprintf(stdout, "%sUpdated: %s%s\n", colorGray, task.UpdatedAt.Format("2006-01-02 15:04:05"), colorReset)

	// Notes
	if len(task.Notes) > 0 {
		fmt.Fprintln(stdout)
		fmt.Fprintf(stdout, "Notes (%d):\n", len(task.Notes))
		for _, note := range task.Notes {
			fmt.Fprintf(stdout, "  %s[%s]%s %s\n",
				colorGray,
				note.CreatedAt.Format("2006-01-02 15:04"),
				colorReset,
				note.Content,
			)
		}
	}

	return nil
}
