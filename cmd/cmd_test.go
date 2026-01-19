package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/jackreid/task/internal/model"
)

// testEnv sets up a test environment with isolated stdout/stderr and temp directory
type testEnv struct {
	origStdout  *os.File
	origStderr  *os.File
	origWorkDir string
	origEditor  string
	origVisual  string
	stdout      *bytes.Buffer
	stderr      *bytes.Buffer
	cleanup     func()
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	env := &testEnv{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}

	// Save original writers
	env.origWorkDir = workDir
	env.origEditor = os.Getenv("EDITOR")
	env.origVisual = os.Getenv("VISUAL")

	// Redirect output
	stdout = env.stdout
	stderr = env.stderr

	// Create temp directory
	tmpDir := t.TempDir()
	workDir = tmpDir
	os.Setenv("EDITOR", "true")
	os.Unsetenv("VISUAL")

	env.cleanup = func() {
		stdout = os.Stdout
		stderr = os.Stderr
		workDir = env.origWorkDir
		if env.origEditor == "" {
			os.Unsetenv("EDITOR")
		} else {
			os.Setenv("EDITOR", env.origEditor)
		}
		if env.origVisual == "" {
			os.Unsetenv("VISUAL")
		} else {
			os.Setenv("VISUAL", env.origVisual)
		}
	}

	return env
}

func TestRunHelp(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	tests := []struct {
		args []string
	}{
		{[]string{}},
		{[]string{"help"}},
		{[]string{"-h"}},
		{[]string{"--help"}},
	}

	for _, tt := range tests {
		env.stdout.Reset()
		err := run(tt.args)
		if err != nil {
			t.Errorf("run(%v) error = %v", tt.args, err)
		}
		if !strings.Contains(env.stdout.String(), "task - a simple task management app") {
			t.Errorf("run(%v) did not output help text", tt.args)
		}
	}
}

func TestRunUnknownCommand(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	err := run([]string{"unknown"})
	if err == nil {
		t.Error("run(unknown) should return error")
	}
}

func TestRunInit(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	err := run([]string{"init"})
	if err != nil {
		t.Errorf("run(init) error = %v", err)
	}

	if !strings.Contains(env.stdout.String(), "Initialized") {
		t.Error("init should print initialization message")
	}

	// Check files were created
	taskFile := workDir + "/.task/task.json"
	if _, err := os.Stat(taskFile); os.IsNotExist(err) {
		t.Error("init did not create .task/task.json")
	}
}

func TestRunInitAlreadyInitialized(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	// Init once
	run([]string{"init"})

	// Init again should fail
	env.stderr.Reset()
	err := run([]string{"init"})
	if err == nil {
		t.Error("second init should return error")
	}
}

func TestRunNew(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})

	env.stdout.Reset()
	err := run([]string{"new", "Test Task"})
	if err != nil {
		t.Errorf("run(new) error = %v", err)
	}

	if !strings.Contains(env.stdout.String(), "Created task") {
		t.Error("new should print creation message")
	}
}

func TestRunNewWithFlags(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})

	// Flags must come before positional arguments for Go's flag package
	err := run([]string{"new", "-t", "bug", "-d", "Description", "-l", "urgent", "-l", "frontend", "Test Bug"})
	if err != nil {
		t.Errorf("run(new with flags) error = %v", err)
	}
}

func TestRunNewInvalidType(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})

	// Flags must come before positional arguments for Go's flag package
	err := run([]string{"new", "-t", "invalid", "Task"})
	if err == nil {
		t.Error("new with invalid type should return error")
	}
}

func TestRunList(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Task 1"})
	run([]string{"new", "-t", "bug", "Task 2"})

	env.stdout.Reset()
	err := run([]string{"list"})
	if err != nil {
		t.Errorf("run(list) error = %v", err)
	}

	output := env.stdout.String()
	if !strings.Contains(output, "Task 1") || !strings.Contains(output, "Task 2") {
		t.Error("list should show all tasks")
	}
}

func TestRunListEmpty(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})

	env.stdout.Reset()
	err := run([]string{"list"})
	if err != nil {
		t.Errorf("run(list) error = %v", err)
	}

	if !strings.Contains(env.stdout.String(), "No tasks found") {
		t.Error("list should show 'No tasks found' when empty")
	}
}

func TestRunListJSON(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Task 1"})

	env.stdout.Reset()
	err := run([]string{"list", "--json"})
	if err != nil {
		t.Errorf("run(list --json) error = %v", err)
	}

	var tasks []model.Task
	if err := json.Unmarshal(env.stdout.Bytes(), &tasks); err != nil {
		t.Errorf("list --json output is not valid JSON: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("list --json returned %d tasks, want 1", len(tasks))
	}
}

func TestRunListWithFilters(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	// Flags must come before positional arguments for Go's flag package
	run([]string{"new", "-t", "task", "Task 1"})
	run([]string{"new", "-t", "bug", "-l", "urgent", "Bug 1"})

	// Test filter by type
	env.stdout.Reset()
	err := run([]string{"list", "-t", "bug"})
	if err != nil {
		t.Errorf("run(list -t bug) error = %v", err)
	}
	output := env.stdout.String()
	if !strings.Contains(output, "Bug 1") {
		t.Errorf("run(list -t bug) should contain 'Bug 1', got: %s", output)
	}
	if strings.Contains(output, "Task 1") {
		t.Errorf("run(list -t bug) should not contain 'Task 1'")
	}

	// Test filter by label
	env.stdout.Reset()
	err = run([]string{"list", "-l", "urgent"})
	if err != nil {
		t.Errorf("run(list -l urgent) error = %v", err)
	}
	output = env.stdout.String()
	if !strings.Contains(output, "Bug 1") {
		t.Errorf("run(list -l urgent) should contain 'Bug 1', got: %s", output)
	}
	if strings.Contains(output, "Task 1") {
		t.Errorf("run(list -l urgent) should not contain 'Task 1'")
	}
}

func TestRunUpdate(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Original Title"})

	// Get the task ID from the output
	output := env.stdout.String()
	taskID := extractTaskID(output)

	env.stdout.Reset()
	err := run([]string{"update", taskID, "-n", "New Title", "-s", "progress"})
	if err != nil {
		t.Errorf("run(update) error = %v", err)
	}

	if !strings.Contains(env.stdout.String(), "Updated task") {
		t.Error("update should print update message")
	}
}

func TestRunUpdateNotFound(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})

	err := run([]string{"update", "xxx", "-n", "New Title"})
	if err == nil {
		t.Error("update with non-existent ID should return error")
	}
}

func TestRunUpdateMissingID(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})

	err := run([]string{"update"})
	if err == nil {
		t.Error("update without ID should return error")
	}
}

func TestRunShow(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	// Flags must come before positional arguments for Go's flag package
	run([]string{"new", "-d", "Description", "-l", "mylabel", "Test Task"})

	output := env.stdout.String()
	taskID := extractTaskID(output)

	env.stdout.Reset()
	err := run([]string{"show", taskID})
	if err != nil {
		t.Errorf("run(show) error = %v", err)
	}

	showOutput := env.stdout.String()
	if !strings.Contains(showOutput, "Test Task") {
		t.Errorf("show should display task title, got: %s", showOutput)
	}
	if !strings.Contains(showOutput, "Description") {
		t.Errorf("show should display description, got: %s", showOutput)
	}
	if !strings.Contains(showOutput, "mylabel") {
		t.Errorf("show should display labels, got: %s", showOutput)
	}
}

func TestRunShowJSON(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Test Task"})

	output := env.stdout.String()
	taskID := extractTaskID(output)

	env.stdout.Reset()
	// Flags must come before positional arguments
	err := run([]string{"show", "--json", taskID})
	if err != nil {
		t.Errorf("run(show --json) error = %v", err)
	}

	// Parse the JSON output (need to handle the struct properly)
	var result map[string]interface{}
	if err := json.Unmarshal(env.stdout.Bytes(), &result); err != nil {
		t.Errorf("show --json output is not valid JSON: %v\nOutput: %s", err, env.stdout.String())
		return
	}

	if result["title"] != "Test Task" {
		t.Errorf("show --json returned title %q, want %q", result["title"], "Test Task")
	}
}

func TestRunShowNotFound(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})

	err := run([]string{"show", "xxx"})
	if err == nil {
		t.Error("show with non-existent ID should return error")
	}
}

func TestRunNote(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Test Task"})

	output := env.stdout.String()
	taskID := extractTaskID(output)

	env.stdout.Reset()
	err := run([]string{"note", taskID, "This is a note"})
	if err != nil {
		t.Errorf("run(note) error = %v", err)
	}

	if !strings.Contains(env.stdout.String(), "Added note") {
		t.Error("note should print confirmation message")
	}

	// Verify note was added
	env.stdout.Reset()
	run([]string{"show", taskID})
	if !strings.Contains(env.stdout.String(), "This is a note") {
		t.Error("note content should be visible in show")
	}
}

func TestRunNoteMissingContent(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Test Task"})

	output := env.stdout.String()
	taskID := extractTaskID(output)

	err := run([]string{"note", taskID})
	if err == nil {
		t.Error("note without content should return error")
	}
}

func TestRunReady(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Todo Task"})
	run([]string{"new", "Progress Task"})

	// Get second task ID and set it to progress
	output := env.stdout.String()
	lines := strings.Split(output, "\n")
	var progressTaskID string
	for _, line := range lines {
		if strings.Contains(line, "Progress Task") {
			progressTaskID = extractTaskID(line)
			break
		}
	}
	if progressTaskID != "" {
		run([]string{"update", progressTaskID, "-s", "progress"})
	}

	env.stdout.Reset()
	err := run([]string{"ready"})
	if err != nil {
		t.Errorf("run(ready) error = %v", err)
	}

	output = env.stdout.String()
	if !strings.Contains(output, "Todo Task") {
		t.Error("ready should show todo tasks")
	}
}

func TestRunTake(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Test Task"})

	output := env.stdout.String()
	taskID := extractTaskID(output)

	env.stdout.Reset()
	err := run([]string{"take", taskID})
	if err != nil {
		t.Errorf("run(take) error = %v", err)
	}

	if !strings.Contains(env.stdout.String(), "progress") {
		t.Error("take should set status to progress")
	}
}

func TestRunComplete(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Test Task"})

	output := env.stdout.String()
	taskID := extractTaskID(output)

	env.stdout.Reset()
	err := run([]string{"complete", taskID})
	if err != nil {
		t.Errorf("run(complete) error = %v", err)
	}

	if !strings.Contains(env.stdout.String(), "done") {
		t.Error("complete should set status to done")
	}
}

func TestRunBlock(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Test Task"})

	output := env.stdout.String()
	taskID := extractTaskID(output)

	env.stdout.Reset()
	err := run([]string{"block", taskID})
	if err != nil {
		t.Errorf("run(block) error = %v", err)
	}

	if !strings.Contains(env.stdout.String(), "blocked") {
		t.Error("block should set status to blocked")
	}
}

func TestRunAbandon(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Test Task"})

	output := env.stdout.String()
	taskID := extractTaskID(output)

	env.stdout.Reset()
	err := run([]string{"abandon", taskID})
	if err != nil {
		t.Errorf("run(abandon) error = %v", err)
	}

	if !strings.Contains(env.stdout.String(), "abandon") {
		t.Error("abandon should set status to abandon")
	}
}

func TestAliasesMissingID(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})

	aliases := []string{"take", "complete", "block", "abandon"}
	for _, alias := range aliases {
		err := run([]string{alias})
		if err == nil {
			t.Errorf("%s without ID should return error", alias)
		}
	}
}

func TestFlexibleFlagOrderNew(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})

	// Test: positional argument before flags
	env.stdout.Reset()
	err := run([]string{"new", "Test Task", "--type", "bug", "--label", "urgent"})
	if err != nil {
		t.Errorf("run(new with positional before flags) error = %v", err)
	}

	output := env.stdout.String()
	if !strings.Contains(output, "Created task") {
		t.Error("new should create task")
	}

	// Verify task was created with correct type and label
	taskID := extractTaskID(output)
	env.stdout.Reset()
	run([]string{"show", taskID, "--json"})
	var task map[string]interface{}
	if err := json.Unmarshal(env.stdout.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}

	if task["type"] != "bug" {
		t.Errorf("task type = %q, want %q", task["type"], "bug")
	}

	labels, ok := task["labels"].([]interface{})
	if !ok || len(labels) != 1 || labels[0] != "urgent" {
		t.Errorf("task labels = %v, want [urgent]", labels)
	}
}

func TestFlexibleFlagOrderNewMultipleFlags(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})

	// Test: positional argument with multiple flags after it
	env.stdout.Reset()
	err := run([]string{"new", "Multi Flag Task", "--type", "feature", "--label", "frontend", "--label", "priority", "--description", "Test description"})
	if err != nil {
		t.Errorf("run(new with multiple flags after positional) error = %v", err)
	}

	taskID := extractTaskID(env.stdout.String())
	env.stdout.Reset()
	run([]string{"show", taskID, "--json"})
	var task map[string]interface{}
	if err := json.Unmarshal(env.stdout.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}

	if task["type"] != "feature" {
		t.Errorf("task type = %q, want %q", task["type"], "feature")
	}

	labels, ok := task["labels"].([]interface{})
	if !ok || len(labels) != 2 {
		t.Errorf("task labels = %v, want 2 labels", labels)
	}

	desc, ok := task["description"].(string)
	if !ok || desc != "Test description" {
		t.Errorf("task description = %q, want %q", desc, "Test description")
	}
}

func TestFlexibleFlagOrderUpdate(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Original Task"})
	taskID := extractTaskID(env.stdout.String())

	// Test: positional argument (task ID) before flags
	env.stdout.Reset()
	err := run([]string{"update", taskID, "--name", "Updated Task", "--status", "progress", "--label", "updated"})
	if err != nil {
		t.Errorf("run(update with positional before flags) error = %v", err)
	}

	env.stdout.Reset()
	run([]string{"show", taskID, "--json"})
	var task map[string]interface{}
	if err := json.Unmarshal(env.stdout.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}

	if task["title"] != "Updated Task" {
		t.Errorf("task title = %q, want %q", task["title"], "Updated Task")
	}

	if task["status"] != "progress" {
		t.Errorf("task status = %q, want %q", task["status"], "progress")
	}
}

func TestFlexibleFlagOrderShow(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Test Show Task"})
	taskID := extractTaskID(env.stdout.String())

	// Test: positional argument (task ID) before flags
	env.stdout.Reset()
	err := run([]string{"show", taskID, "--json"})
	if err != nil {
		t.Errorf("run(show with positional before flags) error = %v", err)
	}

	var task map[string]interface{}
	if err := json.Unmarshal(env.stdout.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}

	if task["title"] != "Test Show Task" {
		t.Errorf("task title = %q, want %q", task["title"], "Test Show Task")
	}
}

func TestFlexibleFlagOrderBackwardCompatibility(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})

	// Test: old format (flags before positional) still works
	env.stdout.Reset()
	err := run([]string{"new", "--type", "bug", "--label", "urgent", "Backward Compat Task"})
	if err != nil {
		t.Errorf("run(new with flags before positional) error = %v", err)
	}

	taskID := extractTaskID(env.stdout.String())
	env.stdout.Reset()
	run([]string{"show", taskID, "--json"})
	var task map[string]interface{}
	if err := json.Unmarshal(env.stdout.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}

	if task["type"] != "bug" {
		t.Errorf("task type = %q, want %q", task["type"], "bug")
	}

	if task["title"] != "Backward Compat Task" {
		t.Errorf("task title = %q, want %q", task["title"], "Backward Compat Task")
	}
}

func TestRunNewWithTitleCreatesImmediately(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})

	// task new "test" should create immediately without launching editor
	env.stdout.Reset()
	err := run([]string{"new", "test"})
	if err != nil {
		t.Errorf("run(new \"test\") error = %v", err)
	}

	if !strings.Contains(env.stdout.String(), "Created task") {
		t.Error("new with title should create task immediately")
	}

	// Verify task was created
	taskID := extractTaskID(env.stdout.String())
	env.stdout.Reset()
	run([]string{"show", taskID, "--json"})
	var task map[string]interface{}
	if err := json.Unmarshal(env.stdout.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}

	if task["title"] != "test" {
		t.Errorf("task title = %q, want %q", task["title"], "test")
	}
}

func TestRunNewLaunchesEditor(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})

	// Create a mock editor script that writes a valid task template
	editorScript := `#!/bin/sh
cat > "$1" << 'EOF'
---
title: "Editor Created Task"
type: task
status: todo
labels: []
---
Description from editor
EOF
`
	tmpEditor := createTempEditorScript(t, editorScript)
	defer os.Remove(tmpEditor)

	// Override EDITOR for this test
	origEditor := os.Getenv("EDITOR")
	os.Setenv("EDITOR", tmpEditor)
	defer func() {
		if origEditor == "" {
			os.Unsetenv("EDITOR")
		} else {
			os.Setenv("EDITOR", origEditor)
		}
	}()

	// task new without arguments should launch editor
	env.stdout.Reset()
	err := run([]string{"new"})
	if err != nil {
		t.Errorf("run(new) without args should launch editor, got error: %v", err)
	}

	// Editor was launched, task should be created with editor content
	if !strings.Contains(env.stdout.String(), "Created task") {
		t.Error("new without args should create task via editor")
	}

	// Verify task has content from editor
	taskID := extractTaskID(env.stdout.String())
	env.stdout.Reset()
	run([]string{"show", taskID, "--json"})
	var task map[string]interface{}
	if err := json.Unmarshal(env.stdout.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}

	if task["title"] != "Editor Created Task" {
		t.Errorf("task title = %q, want %q", task["title"], "Editor Created Task")
	}
}

func TestRunEditLaunchesEditor(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Original Task"})
	taskID := extractTaskID(env.stdout.String())

	// Create a mock editor script that modifies the task
	editorScript := `#!/bin/sh
cat > "$1" << 'EOF'
---
title: "Edited Task"
type: bug
status: progress
labels: [urgent]
---
Updated description
EOF
`
	tmpEditor := createTempEditorScript(t, editorScript)
	defer os.Remove(tmpEditor)

	// Override EDITOR for this test
	origEditor := os.Getenv("EDITOR")
	os.Setenv("EDITOR", tmpEditor)
	defer func() {
		if origEditor == "" {
			os.Unsetenv("EDITOR")
		} else {
			os.Setenv("EDITOR", origEditor)
		}
	}()

	// task edit $id should launch editor
	env.stdout.Reset()
	err := run([]string{"edit", taskID})
	if err != nil {
		t.Errorf("run(edit) should launch editor, got error: %v", err)
	}

	if !strings.Contains(env.stdout.String(), "Updated task") {
		t.Error("edit should update task via editor")
	}

	// Verify task was updated with editor content
	env.stdout.Reset()
	run([]string{"show", taskID, "--json"})
	var task map[string]interface{}
	if err := json.Unmarshal(env.stdout.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}

	if task["title"] != "Edited Task" {
		t.Errorf("task title = %q, want %q", task["title"], "Edited Task")
	}
	if task["type"] != "bug" {
		t.Errorf("task type = %q, want %q", task["type"], "bug")
	}
}

func TestRunEditWithFlagsEditsImmediately(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	run([]string{"init"})
	run([]string{"new", "Original Task"})
	taskID := extractTaskID(env.stdout.String())

	// task edit $id --type bug should edit immediately without launching editor
	env.stdout.Reset()
	err := run([]string{"edit", taskID, "--type", "bug"})
	if err != nil {
		t.Errorf("run(edit with flags) error = %v", err)
	}

	if !strings.Contains(env.stdout.String(), "Updated task") {
		t.Error("edit with flags should update task immediately")
	}

	// Verify task was updated
	env.stdout.Reset()
	run([]string{"show", taskID, "--json"})
	var task map[string]interface{}
	if err := json.Unmarshal(env.stdout.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}

	if task["type"] != "bug" {
		t.Errorf("task type = %q, want %q", task["type"], "bug")
	}
	// Title should remain unchanged
	if task["title"] != "Original Task" {
		t.Errorf("task title = %q, want %q", task["title"], "Original Task")
	}
}

// createTempEditorScript creates a temporary executable script file
func createTempEditorScript(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "task-test-editor-*.sh")
	if err != nil {
		t.Fatalf("failed to create temp editor script: %v", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		t.Fatalf("failed to write editor script: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		t.Fatalf("failed to close editor script: %v", err)
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		t.Fatalf("failed to make editor script executable: %v", err)
	}

	return tmpPath
}

// extractTaskID extracts a task ID from output like "Created task abc: Title"
func extractTaskID(output string) string {
	// Look for "task xxx:" or "task xxx " pattern
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Created task") || strings.Contains(line, "Updated task") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "task" && i+1 < len(parts) {
					id := strings.TrimSuffix(parts[i+1], ":")
					return id
				}
			}
		}
	}
	return ""
}
