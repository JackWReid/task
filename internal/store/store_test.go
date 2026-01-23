package store

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackreid/task/internal/model"
)

func TestNewStore(t *testing.T) {
	s := New("")
	if s.dir != "." {
		t.Errorf("New(\"\").dir = %q, want %q", s.dir, ".")
	}

	s = New("/tmp/test")
	if s.dir != "/tmp/test" {
		t.Errorf("New(\"/tmp/test\").dir = %q, want %q", s.dir, "/tmp/test")
	}
}

func TestStoreInit(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)

	err := s.Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Check directory was created
	taskDir := filepath.Join(tmpDir, TaskDir)
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		t.Error("Init() did not create .task directory")
	}

	// Check file was created
	taskFile := filepath.Join(taskDir, TaskFile)
	if _, err := os.Stat(taskFile); os.IsNotExist(err) {
		t.Error("Init() did not create task.json")
	}

	// Check file is empty (JSONL format with no tasks)
	data, err := os.ReadFile(taskFile)
	if err != nil {
		t.Fatalf("reading task.json: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("task.json should be empty, got %d bytes: %q", len(data), string(data))
	}
}

func TestStoreInitAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)

	// Initialize once
	err := s.Init()
	if err != nil {
		t.Fatalf("First Init() error = %v", err)
	}

	// Try to initialize again
	err = s.Init()
	if err == nil {
		t.Error("Second Init() should return error")
	}
}

func TestStoreIsInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)

	if s.IsInitialized() {
		t.Error("IsInitialized() = true before Init()")
	}

	s.Init()

	if !s.IsInitialized() {
		t.Error("IsInitialized() = false after Init()")
	}
}

func TestStoreLoadEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	tasks, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("Load() returned %d tasks, want 0", len(tasks))
	}
}

func TestStoreLoadNotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)

	_, err := s.Load()
	if err == nil {
		t.Error("Load() should return error when not initialized")
	}
}

func TestStoreSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	task := model.NewTask("abc", "Test Task", model.TypeTask)
	task.SetDescription("A description")
	task.AddLabel("label1")

	err := s.Save([]model.Task{*task})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	tasks, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Load() returned %d tasks, want 1", len(tasks))
	}

	loaded := tasks[0]
	if loaded.ID != task.ID {
		t.Errorf("loaded.ID = %q, want %q", loaded.ID, task.ID)
	}
	if loaded.Title != task.Title {
		t.Errorf("loaded.Title = %q, want %q", loaded.Title, task.Title)
	}
	if loaded.Description == nil || *loaded.Description != *task.Description {
		t.Error("loaded.Description mismatch")
	}
	if len(loaded.Labels) != 1 || loaded.Labels[0] != "label1" {
		t.Errorf("loaded.Labels = %v, want [label1]", loaded.Labels)
	}
}

func TestStoreAdd(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	task := model.NewTask("abc", "Test Task", model.TypeTask)
	err := s.Add(task)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	tasks, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Load() returned %d tasks, want 1", len(tasks))
	}
	if tasks[0].ID != "abc" {
		t.Errorf("tasks[0].ID = %q, want %q", tasks[0].ID, "abc")
	}
}

func TestStoreFindByID(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	task := model.NewTask("abc", "Test Task", model.TypeTask)
	s.Add(task)

	// Find existing task
	found, err := s.FindByID("abc")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found == nil {
		t.Fatal("FindByID() returned nil for existing task")
	}
	if found.ID != "abc" {
		t.Errorf("found.ID = %q, want %q", found.ID, "abc")
	}

	// Find non-existing task
	found, err = s.FindByID("xyz")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found != nil {
		t.Error("FindByID() should return nil for non-existing task")
	}
}

func TestStoreUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	task := model.NewTask("abc", "Test Task", model.TypeTask)
	s.Add(task)

	// Update the task
	task.SetTitle("Updated Title")
	err := s.Update(task)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update
	found, _ := s.FindByID("abc")
	if found.Title != "Updated Title" {
		t.Errorf("found.Title = %q, want %q", found.Title, "Updated Title")
	}
}

func TestStoreUpdateNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	task := model.NewTask("abc", "Test Task", model.TypeTask)
	err := s.Update(task)
	if err == nil {
		t.Error("Update() should return error for non-existing task")
	}
}

func TestStoreGetExistingIDs(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	s.Add(model.NewTask("abc", "Task 1", model.TypeTask))
	s.Add(model.NewTask("xyz", "Task 2", model.TypeTask))

	ids, err := s.GetExistingIDs()
	if err != nil {
		t.Fatalf("GetExistingIDs() error = %v", err)
	}

	if len(ids) != 2 {
		t.Errorf("GetExistingIDs() returned %d IDs, want 2", len(ids))
	}
	if !ids["abc"] {
		t.Error("GetExistingIDs() missing 'abc'")
	}
	if !ids["xyz"] {
		t.Error("GetExistingIDs() missing 'xyz'")
	}
}

func TestStoreListSorted(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	// Create tasks with explicit different updated_at times
	now := time.Now().UTC()
	task1 := model.NewTask("aaa", "Task 1", model.TypeTask)
	task1.CreatedAt = now.Add(-2 * time.Hour)
	task1.UpdatedAt = now.Add(-2 * time.Hour)

	task2 := model.NewTask("bbb", "Task 2", model.TypeTask)
	task2.CreatedAt = now.Add(-1 * time.Hour)
	task2.UpdatedAt = now.Add(-1 * time.Hour)

	task3 := model.NewTask("ccc", "Task 3", model.TypeTask)
	task3.CreatedAt = now
	task3.UpdatedAt = now

	// Save all tasks
	s.Save([]model.Task{*task1, *task2, *task3})

	tasks, err := s.ListSorted()
	if err != nil {
		t.Fatalf("ListSorted() error = %v", err)
	}

	if len(tasks) != 3 {
		t.Fatalf("ListSorted() returned %d tasks, want 3", len(tasks))
	}

	// Should be sorted by UpdatedAt descending (newest first)
	if tasks[0].ID != "ccc" {
		t.Errorf("tasks[0].ID = %q, want %q (newest)", tasks[0].ID, "ccc")
	}
	if tasks[2].ID != "aaa" {
		t.Errorf("tasks[2].ID = %q, want %q (oldest)", tasks[2].ID, "aaa")
	}
}

func TestStoreListFiltered(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	task1 := model.NewTask("aaa", "Task 1", model.TypeTask)
	task1.SetStatus(model.StatusTodo)
	task1.AddLabel("frontend")

	task2 := model.NewTask("bbb", "Task 2", model.TypeBug)
	task2.SetStatus(model.StatusProgress)
	task2.AddLabel("backend")

	task3 := model.NewTask("ccc", "Task 3", model.TypeFeature)
	task3.SetStatus(model.StatusTodo)
	task3.AddLabel("frontend")

	s.Add(task1)
	s.Add(task2)
	s.Add(task3)

	// Filter by status
	statusTodo := model.StatusTodo
	tasks, err := s.ListFiltered(Filter{Status: &statusTodo})
	if err != nil {
		t.Fatalf("ListFiltered() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("ListFiltered(status=todo) returned %d tasks, want 2", len(tasks))
	}

	// Filter by type
	typeBug := model.TypeBug
	tasks, err = s.ListFiltered(Filter{Type: &typeBug})
	if err != nil {
		t.Fatalf("ListFiltered() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("ListFiltered(type=bug) returned %d tasks, want 1", len(tasks))
	}

	// Filter by label
	labelFrontend := "frontend"
	tasks, err = s.ListFiltered(Filter{Label: &labelFrontend})
	if err != nil {
		t.Fatalf("ListFiltered() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("ListFiltered(label=frontend) returned %d tasks, want 2", len(tasks))
	}

	// Combined filter
	tasks, err = s.ListFiltered(Filter{Status: &statusTodo, Label: &labelFrontend})
	if err != nil {
		t.Fatalf("ListFiltered() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("ListFiltered(status=todo, label=frontend) returned %d tasks, want 2", len(tasks))
	}
}

func TestStoreListFilteredEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	task := model.NewTask("abc", "Test Task", model.TypeTask)
	s.Add(task)

	// No filter - should return all
	tasks, err := s.ListFiltered(Filter{})
	if err != nil {
		t.Fatalf("ListFiltered() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("ListFiltered({}) returned %d tasks, want 1", len(tasks))
	}
}

func TestStoreEnsuresNonNilSlices(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	// Write JSON with null values
	taskFile := filepath.Join(tmpDir, TaskDir, TaskFile)
	data := []byte(`[{"id":"abc","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z","title":"Test","description":null,"type":"task","status":"todo","labels":null,"notes":null}]`)
	os.WriteFile(taskFile, data, 0644)

	tasks, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if tasks[0].Labels == nil {
		t.Error("Labels should not be nil after Load()")
	}
	if tasks[0].Notes == nil {
		t.Error("Notes should not be nil after Load()")
	}
}

func TestStoreDelete(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	task1 := model.NewTask("aaa", "Task 1", model.TypeTask)
	task2 := model.NewTask("bbb", "Task 2", model.TypeTask)
	task3 := model.NewTask("ccc", "Task 3", model.TypeTask)

	s.Add(task1)
	s.Add(task2)
	s.Add(task3)

	// Delete middle task
	err := s.Delete("bbb")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	tasks, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("After Delete(), Load() returned %d tasks, want 2", len(tasks))
	}

	// Verify the correct task was deleted
	found := false
	for _, task := range tasks {
		if task.ID == "bbb" {
			found = true
		}
	}
	if found {
		t.Error("Deleted task 'bbb' should not be in store")
	}

	// Verify remaining tasks are still there
	foundA := false
	foundC := false
	for _, task := range tasks {
		if task.ID == "aaa" {
			foundA = true
		}
		if task.ID == "ccc" {
			foundC = true
		}
	}
	if !foundA {
		t.Error("Task 'aaa' should still be in store")
	}
	if !foundC {
		t.Error("Task 'ccc' should still be in store")
	}
}

func TestStoreDeleteNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	task := model.NewTask("abc", "Task 1", model.TypeTask)
	s.Add(task)

	err := s.Delete("xyz")
	if err == nil {
		t.Error("Delete() should return error for non-existent task")
	}

	// Verify original task still exists
	tasks, _ := s.Load()
	if len(tasks) != 1 {
		t.Errorf("After failed Delete(), Load() returned %d tasks, want 1", len(tasks))
	}
}

func TestStoreDeleteLast(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	task := model.NewTask("abc", "Task 1", model.TypeTask)
	s.Add(task)

	err := s.Delete("abc")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	tasks, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("After deleting last task, Load() returned %d tasks, want 0", len(tasks))
	}
}

func TestStoreClean(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	task1 := model.NewTask("aaa", "Todo Task", model.TypeTask)
	task1.SetStatus(model.StatusTodo)

	task2 := model.NewTask("bbb", "Done Task", model.TypeTask)
	task2.SetStatus(model.StatusDone)

	task3 := model.NewTask("ccc", "Abandon Task", model.TypeTask)
	task3.SetStatus(model.StatusAbandon)

	task4 := model.NewTask("ddd", "Progress Task", model.TypeTask)
	task4.SetStatus(model.StatusProgress)

	task5 := model.NewTask("eee", "Blocked Task", model.TypeTask)
	task5.SetStatus(model.StatusBlocked)

	s.Add(task1)
	s.Add(task2)
	s.Add(task3)
	s.Add(task4)
	s.Add(task5)

	deleted, err := s.Clean()
	if err != nil {
		t.Fatalf("Clean() error = %v", err)
	}

	if deleted != 2 {
		t.Errorf("Clean() deleted %d tasks, want 2", deleted)
	}

	tasks, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("After Clean(), Load() returned %d tasks, want 3", len(tasks))
	}

	// Verify done and abandon tasks are gone
	for _, task := range tasks {
		if task.Status == model.StatusDone || task.Status == model.StatusAbandon {
			t.Errorf("Task %s with status %s should have been deleted", task.ID, task.Status)
		}
	}

	// Verify remaining tasks are correct
	expectedIDs := map[string]bool{"aaa": false, "ddd": false, "eee": false}
	for _, task := range tasks {
		if _, ok := expectedIDs[task.ID]; ok {
			expectedIDs[task.ID] = true
		}
	}
	for id, found := range expectedIDs {
		if !found {
			t.Errorf("Task %s should still be in store", id)
		}
	}
}

func TestStoreCleanNoClosedTasks(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	task1 := model.NewTask("aaa", "Todo Task", model.TypeTask)
	task1.SetStatus(model.StatusTodo)

	task2 := model.NewTask("bbb", "Progress Task", model.TypeTask)
	task2.SetStatus(model.StatusProgress)

	s.Add(task1)
	s.Add(task2)

	deleted, err := s.Clean()
	if err != nil {
		t.Fatalf("Clean() error = %v", err)
	}

	if deleted != 0 {
		t.Errorf("Clean() deleted %d tasks, want 0", deleted)
	}

	tasks, _ := s.Load()
	if len(tasks) != 2 {
		t.Errorf("After Clean(), Load() returned %d tasks, want 2", len(tasks))
	}
}

func TestStoreCleanEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	deleted, err := s.Clean()
	if err != nil {
		t.Fatalf("Clean() error = %v", err)
	}

	if deleted != 0 {
		t.Errorf("Clean() deleted %d tasks, want 0", deleted)
	}
}

func TestStoreCleanAll(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	task1 := model.NewTask("aaa", "Done Task", model.TypeTask)
	task1.SetStatus(model.StatusDone)

	task2 := model.NewTask("bbb", "Abandon Task", model.TypeTask)
	task2.SetStatus(model.StatusAbandon)

	s.Add(task1)
	s.Add(task2)

	deleted, err := s.Clean()
	if err != nil {
		t.Fatalf("Clean() error = %v", err)
	}

	if deleted != 2 {
		t.Errorf("Clean() deleted %d tasks, want 2", deleted)
	}

	tasks, _ := s.Load()
	if len(tasks) != 0 {
		t.Errorf("After Clean(), Load() returned %d tasks, want 0", len(tasks))
	}
}

func TestStoreLoadLegacyJSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	// Write legacy JSON array format
	taskFile := filepath.Join(tmpDir, TaskDir, TaskFile)
	legacyData := []byte(`[
  {
    "id": "abc",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "title": "Legacy Task",
    "description": "From JSON array",
    "type": "task",
    "status": "todo",
    "labels": ["legacy"],
    "notes": []
  },
  {
    "id": "def",
    "created_at": "2024-01-02T00:00:00Z",
    "updated_at": "2024-01-02T00:00:00Z",
    "title": "Another Legacy Task",
    "description": null,
    "type": "bug",
    "status": "progress",
    "labels": [],
    "notes": []
  }
]`)
	if err := os.WriteFile(taskFile, legacyData, 0644); err != nil {
		t.Fatalf("failed to write legacy format: %v", err)
	}

	// Load should work with legacy format
	tasks, err := s.Load()
	if err != nil {
		t.Fatalf("Load() with legacy format error = %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("Load() returned %d tasks, want 2", len(tasks))
	}

	if tasks[0].ID != "abc" || tasks[0].Title != "Legacy Task" {
		t.Errorf("task[0] = {ID: %q, Title: %q}, want {ID: %q, Title: %q}",
			tasks[0].ID, tasks[0].Title, "abc", "Legacy Task")
	}

	if tasks[1].ID != "def" || tasks[1].Title != "Another Legacy Task" {
		t.Errorf("task[1] = {ID: %q, Title: %q}, want {ID: %q, Title: %q}",
			tasks[1].ID, tasks[1].Title, "def", "Another Legacy Task")
	}

	// Save should convert to JSONL format
	if err := s.Save(tasks); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Read the file and verify it's now JSONL
	data, err := os.ReadFile(taskFile)
	if err != nil {
		t.Fatalf("reading task file: %v", err)
	}

	lines := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
	if len(lines) != 2 {
		t.Errorf("JSONL file has %d lines, want 2", len(lines))
	}

	// Each line should be valid JSON
	for i, line := range lines {
		var task model.Task
		if err := json.Unmarshal(line, &task); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}
}

func TestStoreLoadJSONLFormat(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.Init()

	// Write JSONL format
	taskFile := filepath.Join(tmpDir, TaskDir, TaskFile)
	jsonlData := []byte(`{"id":"abc","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z","title":"JSONL Task 1","description":"First task","type":"task","status":"todo","labels":["test"],"notes":[]}
{"id":"def","created_at":"2024-01-02T00:00:00Z","updated_at":"2024-01-02T00:00:00Z","title":"JSONL Task 2","description":null,"type":"bug","status":"progress","labels":[],"notes":[]}
`)
	if err := os.WriteFile(taskFile, jsonlData, 0644); err != nil {
		t.Fatalf("failed to write JSONL format: %v", err)
	}

	// Load should work with JSONL format
	tasks, err := s.Load()
	if err != nil {
		t.Fatalf("Load() with JSONL format error = %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("Load() returned %d tasks, want 2", len(tasks))
	}

	if tasks[0].ID != "abc" || tasks[0].Title != "JSONL Task 1" {
		t.Errorf("task[0] = {ID: %q, Title: %q}, want {ID: %q, Title: %q}",
			tasks[0].ID, tasks[0].Title, "abc", "JSONL Task 1")
	}

	if tasks[1].ID != "def" || tasks[1].Title != "JSONL Task 2" {
		t.Errorf("task[1] = {ID: %q, Title: %q}, want {ID: %q, Title: %q}",
			tasks[1].ID, tasks[1].Title, "def", "JSONL Task 2")
	}
}
