package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/jackreid/task/internal/model"
	"github.com/jackreid/task/internal/store"
)

// ANSI color codes
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
)

func runList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var jsonOutput bool
	var labelFilter string
	var typeFilter string
	var statusFilter string

	fs.BoolVar(&jsonOutput, "json", false, "Output as JSON")
	fs.StringVar(&labelFilter, "l", "", "Filter by label")
	fs.StringVar(&labelFilter, "label", "", "Filter by label")
	fs.StringVar(&typeFilter, "t", "", "Filter by type (task, bug, feature)")
	fs.StringVar(&typeFilter, "type", "", "Filter by type (task, bug, feature)")
	fs.StringVar(&statusFilter, "s", "", "Filter by status (todo, progress, blocked, abandon, done)")
	fs.StringVar(&statusFilter, "status", "", "Filter by status (todo, progress, blocked, abandon, done)")

	fs.Usage = func() {
		fmt.Fprintln(stderr, `List all tasks.

Usage:
  task list [flags]

Flags:
  --json              Output as JSON
  -l, --label string  Filter by label
  -t, --type string   Filter by type: task, bug, feature
  -s, --status string Filter by status: todo, progress, blocked, abandon, done

Examples:
  task list
  task list --json
  task list -s todo
  task list -t bug -l urgent`)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	s := getStore()

	// Build filter
	filter := store.Filter{}

	if statusFilter != "" {
		status, err := model.ParseStatus(statusFilter)
		if err != nil {
			errorf("Error: %v", err)
			return err
		}
		filter.Status = &status
	}

	if typeFilter != "" {
		tt, err := model.ParseTaskType(typeFilter)
		if err != nil {
			errorf("Error: %v", err)
			return err
		}
		filter.Type = &tt
	}

	if labelFilter != "" {
		filter.Label = &labelFilter
	}

	tasks, err := s.ListFiltered(filter)
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	if jsonOutput {
		return printTasksJSON(tasks)
	}

	return printTasksPretty(tasks)
}

func printTasksJSON(tasks []model.Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, string(data))
	return nil
}

func printTasksPretty(tasks []model.Task) error {
	if len(tasks) == 0 {
		fmt.Fprintln(stdout, "No tasks found.")
		return nil
	}

	for _, t := range tasks {
		printTaskLine(t)
	}
	return nil
}

func printTaskLine(t model.Task) {
	statusColor := getStatusColor(t.Status)
	typeIcon := getTypeIcon(t.Type)

	// Format: [ID] [status] [type icon] Title [labels]
	fmt.Fprintf(stdout, "%s%s%s %s%s%s %s %s",
		colorCyan, t.ID, colorReset,
		statusColor, statusSymbol(t.Status), colorReset,
		typeIcon,
		t.Title,
	)

	if len(t.Labels) > 0 {
		fmt.Fprintf(stdout, " %s[%s]%s", colorGray, strings.Join(t.Labels, ", "), colorReset)
	}

	fmt.Fprintln(stdout)
}

func getStatusColor(s model.Status) string {
	switch s {
	case model.StatusTodo:
		return colorYellow
	case model.StatusProgress:
		return colorBlue
	case model.StatusBlocked:
		return colorRed
	case model.StatusAbandon:
		return colorGray
	case model.StatusDone:
		return colorGreen
	default:
		return colorReset
	}
}

func statusSymbol(s model.Status) string {
	switch s {
	case model.StatusTodo:
		return "○"
	case model.StatusProgress:
		return "◐"
	case model.StatusBlocked:
		return "✕"
	case model.StatusAbandon:
		return "⊘"
	case model.StatusDone:
		return "●"
	default:
		return "?"
	}
}

func getTypeIcon(t model.TaskType) string {
	switch t {
	case model.TypeTask:
		return "📋"
	case model.TypeBug:
		return "🐛"
	case model.TypeFeature:
		return "✨"
	default:
		return "📋"
	}
}
