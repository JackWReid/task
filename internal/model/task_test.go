package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestStatusIsValid(t *testing.T) {
	tests := []struct {
		status Status
		valid  bool
	}{
		{StatusTodo, true},
		{StatusProgress, true},
		{StatusBlocked, true},
		{StatusAbandon, true},
		{StatusDone, true},
		{Status("invalid"), false},
		{Status(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.valid {
				t.Errorf("Status(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
			}
		})
	}
}

func TestParseStatus(t *testing.T) {
	tests := []struct {
		input   string
		want    Status
		wantErr bool
	}{
		{"todo", StatusTodo, false},
		{"progress", StatusProgress, false},
		{"blocked", StatusBlocked, false},
		{"abandon", StatusAbandon, false},
		{"done", StatusDone, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseStatus(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStatus(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseStatus(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTaskTypeIsValid(t *testing.T) {
	tests := []struct {
		taskType TaskType
		valid    bool
	}{
		{TypeTask, true},
		{TypeBug, true},
		{TypeFeature, true},
		{TaskType("invalid"), false},
		{TaskType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.taskType), func(t *testing.T) {
			if got := tt.taskType.IsValid(); got != tt.valid {
				t.Errorf("TaskType(%q).IsValid() = %v, want %v", tt.taskType, got, tt.valid)
			}
		})
	}
}

func TestParseTaskType(t *testing.T) {
	tests := []struct {
		input   string
		want    TaskType
		wantErr bool
	}{
		{"task", TypeTask, false},
		{"bug", TypeBug, false},
		{"feature", TypeFeature, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseTaskType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTaskType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseTaskType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewTask(t *testing.T) {
	task := NewTask("abc", "Test Task", TypeTask)

	if task.ID != "abc" {
		t.Errorf("NewTask ID = %q, want %q", task.ID, "abc")
	}
	if task.Title != "Test Task" {
		t.Errorf("NewTask Title = %q, want %q", task.Title, "Test Task")
	}
	if task.Type != TypeTask {
		t.Errorf("NewTask Type = %q, want %q", task.Type, TypeTask)
	}
	if task.Status != StatusTodo {
		t.Errorf("NewTask Status = %q, want %q", task.Status, StatusTodo)
	}
	if task.Description != nil {
		t.Errorf("NewTask Description = %v, want nil", task.Description)
	}
	if len(task.Labels) != 0 {
		t.Errorf("NewTask Labels = %v, want empty slice", task.Labels)
	}
	if len(task.Notes) != 0 {
		t.Errorf("NewTask Notes = %v, want empty slice", task.Notes)
	}
	if task.CreatedAt.IsZero() {
		t.Error("NewTask CreatedAt should not be zero")
	}
	if task.UpdatedAt.IsZero() {
		t.Error("NewTask UpdatedAt should not be zero")
	}
}

func TestTaskSetDescription(t *testing.T) {
	task := NewTask("abc", "Test Task", TypeTask)
	originalUpdatedAt := task.UpdatedAt

	time.Sleep(time.Millisecond)
	task.SetDescription("A description")

	if task.Description == nil {
		t.Fatal("Description should not be nil after SetDescription")
	}
	if *task.Description != "A description" {
		t.Errorf("Description = %q, want %q", *task.Description, "A description")
	}
	if !task.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated after SetDescription")
	}
}

func TestTaskAddLabel(t *testing.T) {
	task := NewTask("abc", "Test Task", TypeTask)

	task.AddLabel("label1")
	if len(task.Labels) != 1 || task.Labels[0] != "label1" {
		t.Errorf("Labels = %v, want [label1]", task.Labels)
	}

	// Adding same label should not duplicate
	task.AddLabel("label1")
	if len(task.Labels) != 1 {
		t.Errorf("Labels = %v, want [label1] (no duplicates)", task.Labels)
	}

	task.AddLabel("label2")
	if len(task.Labels) != 2 {
		t.Errorf("Labels = %v, want [label1, label2]", task.Labels)
	}
}

func TestTaskSetLabels(t *testing.T) {
	task := NewTask("abc", "Test Task", TypeTask)
	task.AddLabel("old")

	task.SetLabels([]string{"new1", "new2"})
	if len(task.Labels) != 2 || task.Labels[0] != "new1" || task.Labels[1] != "new2" {
		t.Errorf("Labels = %v, want [new1, new2]", task.Labels)
	}
}

func TestTaskSetStatus(t *testing.T) {
	task := NewTask("abc", "Test Task", TypeTask)

	err := task.SetStatus(StatusProgress)
	if err != nil {
		t.Errorf("SetStatus(progress) error = %v", err)
	}
	if task.Status != StatusProgress {
		t.Errorf("Status = %q, want %q", task.Status, StatusProgress)
	}

	err = task.SetStatus(Status("invalid"))
	if err == nil {
		t.Error("SetStatus(invalid) should return error")
	}
}

func TestTaskSetType(t *testing.T) {
	task := NewTask("abc", "Test Task", TypeTask)

	err := task.SetType(TypeBug)
	if err != nil {
		t.Errorf("SetType(bug) error = %v", err)
	}
	if task.Type != TypeBug {
		t.Errorf("Type = %q, want %q", task.Type, TypeBug)
	}

	err = task.SetType(TaskType("invalid"))
	if err == nil {
		t.Error("SetType(invalid) should return error")
	}
}

func TestTaskSetTitle(t *testing.T) {
	task := NewTask("abc", "Test Task", TypeTask)
	task.SetTitle("New Title")

	if task.Title != "New Title" {
		t.Errorf("Title = %q, want %q", task.Title, "New Title")
	}
}

func TestTaskAddNote(t *testing.T) {
	task := NewTask("abc", "Test Task", TypeTask)
	originalUpdatedAt := task.UpdatedAt

	time.Sleep(time.Millisecond)
	task.AddNote("abc-123", "This is a note")

	if len(task.Notes) != 1 {
		t.Fatalf("Notes length = %d, want 1", len(task.Notes))
	}
	note := task.Notes[0]
	if note.ID != "abc-123" {
		t.Errorf("Note ID = %q, want %q", note.ID, "abc-123")
	}
	if note.Content != "This is a note" {
		t.Errorf("Note Content = %q, want %q", note.Content, "This is a note")
	}
	if !task.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated after AddNote")
	}
}

func TestTaskHasLabel(t *testing.T) {
	task := NewTask("abc", "Test Task", TypeTask)
	task.AddLabel("label1")

	if !task.HasLabel("label1") {
		t.Error("HasLabel(label1) = false, want true")
	}
	if task.HasLabel("label2") {
		t.Error("HasLabel(label2) = true, want false")
	}
}

func TestTaskJSONRoundTrip(t *testing.T) {
	task := NewTask("abc", "Test Task", TypeBug)
	task.SetDescription("A description")
	task.AddLabel("label1")
	task.AddNote("abc-123", "A note")

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}

	var decoded Task
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}

	if decoded.ID != task.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, task.ID)
	}
	if decoded.Title != task.Title {
		t.Errorf("Title = %q, want %q", decoded.Title, task.Title)
	}
	if decoded.Type != task.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, task.Type)
	}
	if decoded.Status != task.Status {
		t.Errorf("Status = %q, want %q", decoded.Status, task.Status)
	}
	if decoded.Description == nil || *decoded.Description != *task.Description {
		t.Errorf("Description mismatch")
	}
	if len(decoded.Labels) != len(task.Labels) {
		t.Errorf("Labels length = %d, want %d", len(decoded.Labels), len(task.Labels))
	}
	if len(decoded.Notes) != len(task.Notes) {
		t.Errorf("Notes length = %d, want %d", len(decoded.Notes), len(task.Notes))
	}
}

func TestTaskJSONFormat(t *testing.T) {
	task := NewTask("9nk", "Task title", TypeTask)

	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}

	// Verify the JSON structure matches expected format
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}

	expectedKeys := []string{"id", "created_at", "updated_at", "title", "description", "type", "status", "labels", "notes"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("JSON missing key %q", key)
		}
	}

	// Verify datetime format (RFC3339)
	createdAt := m["created_at"].(string)
	_, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		t.Errorf("created_at not in RFC3339 format: %v", err)
	}
}

func TestAllStatuses(t *testing.T) {
	statuses := AllStatuses()
	if len(statuses) != 5 {
		t.Errorf("AllStatuses() length = %d, want 5", len(statuses))
	}
}

func TestAllTaskTypes(t *testing.T) {
	types := AllTaskTypes()
	if len(types) != 3 {
		t.Errorf("AllTaskTypes() length = %d, want 3", len(types))
	}
}

func TestStatusString(t *testing.T) {
	if StatusTodo.String() != "todo" {
		t.Errorf("StatusTodo.String() = %q, want %q", StatusTodo.String(), "todo")
	}
}

func TestTaskTypeString(t *testing.T) {
	if TypeTask.String() != "task" {
		t.Errorf("TypeTask.String() = %q, want %q", TypeTask.String(), "task")
	}
}

func TestNoteJSONRoundTrip(t *testing.T) {
	note := Note{
		ID:        "abc-123",
		CreatedAt: time.Now().UTC(),
		Content:   "Test content",
	}

	data, err := json.Marshal(note)
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}

	var decoded Note
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}

	if decoded.ID != note.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, note.ID)
	}
	if decoded.Content != note.Content {
		t.Errorf("Content = %q, want %q", decoded.Content, note.Content)
	}
	// Time comparison within a second (parsing loses sub-second precision)
	if decoded.CreatedAt.Unix() != note.CreatedAt.Unix() {
		t.Errorf("CreatedAt mismatch")
	}
}
