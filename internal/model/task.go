package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// Status represents the status of a task
type Status string

const (
	StatusTodo     Status = "todo"
	StatusProgress Status = "progress"
	StatusBlocked  Status = "blocked"
	StatusAbandon  Status = "abandon"
	StatusDone     Status = "done"
)

// AllStatuses returns all valid status values
func AllStatuses() []Status {
	return []Status{StatusTodo, StatusProgress, StatusBlocked, StatusAbandon, StatusDone}
}

// IsValid checks if the status is a valid value
func (s Status) IsValid() bool {
	switch s {
	case StatusTodo, StatusProgress, StatusBlocked, StatusAbandon, StatusDone:
		return true
	}
	return false
}

// String returns the string representation of the status
func (s Status) String() string {
	return string(s)
}

// ParseStatus parses a string into a Status
func ParseStatus(s string) (Status, error) {
	status := Status(s)
	if !status.IsValid() {
		return "", fmt.Errorf("invalid status: %s (valid: todo, progress, blocked, abandon, done)", s)
	}
	return status, nil
}

// TaskType represents the type of a task
type TaskType string

const (
	TypeTask    TaskType = "task"
	TypeBug     TaskType = "bug"
	TypeFeature TaskType = "feature"
)

// AllTaskTypes returns all valid task type values
func AllTaskTypes() []TaskType {
	return []TaskType{TypeTask, TypeBug, TypeFeature}
}

// IsValid checks if the task type is a valid value
func (t TaskType) IsValid() bool {
	switch t {
	case TypeTask, TypeBug, TypeFeature:
		return true
	}
	return false
}

// String returns the string representation of the task type
func (t TaskType) String() string {
	return string(t)
}

// ParseTaskType parses a string into a TaskType
func ParseTaskType(s string) (TaskType, error) {
	taskType := TaskType(s)
	if !taskType.IsValid() {
		return "", fmt.Errorf("invalid type: %s (valid: task, bug, feature)", s)
	}
	return taskType, nil
}

// Note represents a note attached to a task
type Note struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Content   string    `json:"content"`
}

// Task represents a task in the system
type Task struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Type        TaskType  `json:"type"`
	Status      Status    `json:"status"`
	Labels      []string  `json:"labels"`
	Notes       []Note    `json:"notes"`
}

// NewTask creates a new task with the given title
func NewTask(id, title string, taskType TaskType) *Task {
	now := time.Now().UTC()
	return &Task{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		Title:     title,
		Type:      taskType,
		Status:    StatusTodo,
		Labels:    []string{},
		Notes:     []Note{},
	}
}

// SetDescription sets the description of the task
func (t *Task) SetDescription(desc string) {
	t.Description = &desc
	t.UpdatedAt = time.Now().UTC()
}

// AddLabel adds a label to the task if not already present
func (t *Task) AddLabel(label string) {
	for _, l := range t.Labels {
		if l == label {
			return
		}
	}
	t.Labels = append(t.Labels, label)
	t.UpdatedAt = time.Now().UTC()
}

// SetLabels replaces the task's labels
func (t *Task) SetLabels(labels []string) {
	t.Labels = labels
	t.UpdatedAt = time.Now().UTC()
}

// SetStatus sets the status of the task
func (t *Task) SetStatus(status Status) error {
	if !status.IsValid() {
		return fmt.Errorf("invalid status: %s", status)
	}
	t.Status = status
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// SetType sets the type of the task
func (t *Task) SetType(taskType TaskType) error {
	if !taskType.IsValid() {
		return fmt.Errorf("invalid type: %s", taskType)
	}
	t.Type = taskType
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// SetTitle sets the title of the task
func (t *Task) SetTitle(title string) {
	t.Title = title
	t.UpdatedAt = time.Now().UTC()
}

// AddNote adds a note to the task
func (t *Task) AddNote(noteID, content string) {
	note := Note{
		ID:        noteID,
		CreatedAt: time.Now().UTC(),
		Content:   content,
	}
	t.Notes = append(t.Notes, note)
	t.UpdatedAt = time.Now().UTC()
}

// HasLabel checks if the task has a specific label
func (t *Task) HasLabel(label string) bool {
	for _, l := range t.Labels {
		if l == label {
			return true
		}
	}
	return false
}

// MarshalJSON implements custom JSON marshaling
func (t Task) MarshalJSON() ([]byte, error) {
	type Alias Task
	return json.Marshal(&struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		*Alias
	}{
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
		Alias:     (*Alias)(&t),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling
func (t *Task) UnmarshalJSON(data []byte) error {
	type Alias Task
	aux := &struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	t.CreatedAt, err = time.Parse(time.RFC3339, aux.CreatedAt)
	if err != nil {
		return fmt.Errorf("parsing created_at: %w", err)
	}
	t.UpdatedAt, err = time.Parse(time.RFC3339, aux.UpdatedAt)
	if err != nil {
		return fmt.Errorf("parsing updated_at: %w", err)
	}
	// Ensure labels and notes are not nil
	if t.Labels == nil {
		t.Labels = []string{}
	}
	if t.Notes == nil {
		t.Notes = []Note{}
	}
	return nil
}

// MarshalJSON implements custom JSON marshaling for Note
func (n Note) MarshalJSON() ([]byte, error) {
	type Alias Note
	return json.Marshal(&struct {
		CreatedAt string `json:"created_at"`
		*Alias
	}{
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
		Alias:     (*Alias)(&n),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Note
func (n *Note) UnmarshalJSON(data []byte) error {
	type Alias Note
	aux := &struct {
		CreatedAt string `json:"created_at"`
		*Alias
	}{
		Alias: (*Alias)(n),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	n.CreatedAt, err = time.Parse(time.RFC3339, aux.CreatedAt)
	if err != nil {
		return fmt.Errorf("parsing created_at: %w", err)
	}
	return nil
}
