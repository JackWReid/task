package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackreid/task/internal/model"
)

const (
	// TaskDir is the directory name for task storage
	TaskDir = ".task"
	// TaskFile is the filename for task storage within TaskDir
	TaskFile = "task.json"
)

// Store handles persistence of tasks to the filesystem
type Store struct {
	dir string
}

// New creates a new Store with the given base directory
// If dir is empty, the current working directory is used
func New(dir string) *Store {
	if dir == "" {
		dir = "."
	}
	return &Store{dir: dir}
}

// taskDir returns the full path to the .task directory
func (s *Store) taskDir() string {
	return filepath.Join(s.dir, TaskDir)
}

// taskFile returns the full path to the task.json file
func (s *Store) taskFile() string {
	return filepath.Join(s.taskDir(), TaskFile)
}

// Init creates the .task directory and empty task.json file
func (s *Store) Init() error {
	taskDir := s.taskDir()

	// Check if already initialized
	if _, err := os.Stat(taskDir); err == nil {
		return fmt.Errorf("task directory already exists: %s", taskDir)
	}

	// Create the .task directory
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return fmt.Errorf("creating task directory: %w", err)
	}

	// Create empty task.json with empty array
	if err := s.Save([]model.Task{}); err != nil {
		// Clean up the directory if we fail to create the file
		os.RemoveAll(taskDir)
		return fmt.Errorf("creating task file: %w", err)
	}

	return nil
}

// IsInitialized checks if the task directory has been initialized
func (s *Store) IsInitialized() bool {
	_, err := os.Stat(s.taskFile())
	return err == nil
}

// Load reads and returns all tasks from the store
func (s *Store) Load() ([]model.Task, error) {
	data, err := os.ReadFile(s.taskFile())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("task not initialized, run 'task init' first")
		}
		return nil, fmt.Errorf("reading task file: %w", err)
	}

	var tasks []model.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("parsing task file: %w", err)
	}

	// Ensure no nil slices
	for i := range tasks {
		if tasks[i].Labels == nil {
			tasks[i].Labels = []string{}
		}
		if tasks[i].Notes == nil {
			tasks[i].Notes = []model.Note{}
		}
	}

	return tasks, nil
}

// Save writes all tasks to the store
func (s *Store) Save(tasks []model.Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding tasks: %w", err)
	}

	if err := os.WriteFile(s.taskFile(), data, 0644); err != nil {
		return fmt.Errorf("writing task file: %w", err)
	}

	return nil
}

// FindByID finds a task by its ID, returns nil if not found
func (s *Store) FindByID(id string) (*model.Task, error) {
	tasks, err := s.Load()
	if err != nil {
		return nil, err
	}

	for i := range tasks {
		if tasks[i].ID == id {
			return &tasks[i], nil
		}
	}

	return nil, nil
}

// Add adds a new task to the store
func (s *Store) Add(task *model.Task) error {
	tasks, err := s.Load()
	if err != nil {
		return err
	}

	tasks = append(tasks, *task)
	return s.Save(tasks)
}

// Update updates an existing task in the store
func (s *Store) Update(task *model.Task) error {
	tasks, err := s.Load()
	if err != nil {
		return err
	}

	found := false
	for i := range tasks {
		if tasks[i].ID == task.ID {
			tasks[i] = *task
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	return s.Save(tasks)
}

// GetExistingIDs returns a map of all existing task IDs
func (s *Store) GetExistingIDs() (map[string]bool, error) {
	tasks, err := s.Load()
	if err != nil {
		return nil, err
	}

	ids := make(map[string]bool)
	for _, t := range tasks {
		ids[t.ID] = true
	}
	return ids, nil
}

// ListSorted returns all tasks sorted by UpdatedAt descending
func (s *Store) ListSorted() ([]model.Task, error) {
	tasks, err := s.Load()
	if err != nil {
		return nil, err
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})

	return tasks, nil
}

// Filter represents filtering options for listing tasks
type Filter struct {
	Status *model.Status
	Type   *model.TaskType
	Label  *string
}

// ListFiltered returns tasks matching the given filter, sorted by UpdatedAt descending
func (s *Store) ListFiltered(filter Filter) ([]model.Task, error) {
	tasks, err := s.ListSorted()
	if err != nil {
		return nil, err
	}

	if filter.Status == nil && filter.Type == nil && filter.Label == nil {
		return tasks, nil
	}

	var result []model.Task
	for _, t := range tasks {
		if filter.Status != nil && t.Status != *filter.Status {
			continue
		}
		if filter.Type != nil && t.Type != *filter.Type {
			continue
		}
		if filter.Label != nil && !t.HasLabel(*filter.Label) {
			continue
		}
		result = append(result, t)
	}

	return result, nil
}

// Delete removes a task from the store by ID
func (s *Store) Delete(id string) error {
	tasks, err := s.Load()
	if err != nil {
		return err
	}

	found := false
	newTasks := make([]model.Task, 0, len(tasks)-1)
	for i := range tasks {
		if tasks[i].ID == id {
			found = true
			continue
		}
		newTasks = append(newTasks, tasks[i])
	}

	if !found {
		return fmt.Errorf("task not found: %s", id)
	}

	return s.Save(newTasks)
}

// Clean removes all closed tasks from the store
// Closed tasks are those with status 'done' or 'abandon'
// Returns the number of tasks deleted
func (s *Store) Clean() (int, error) {
	tasks, err := s.Load()
	if err != nil {
		return 0, err
	}

	deleted := 0
	newTasks := make([]model.Task, 0, len(tasks))
	for i := range tasks {
		if tasks[i].Status == model.StatusDone || tasks[i].Status == model.StatusAbandon {
			deleted++
			continue
		}
		newTasks = append(newTasks, tasks[i])
	}

	if deleted == 0 {
		return 0, nil
	}

	return deleted, s.Save(newTasks)
}
