// Package store provides persistent JSON-based storage for tasks.
// Uses a read-write mutex for concurrent safety — a significant
// performance improvement over Java's synchronized blocks.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Krishnadeshpande2907/TaskSchedulerTUI/internal/model"
)

// Store manages task persistence to a JSON file.
type Store struct {
	mu       sync.RWMutex
	tasks    []model.Task
	filePath string
}

// New creates or loads a Store backed by a JSON file in the user's config directory.
func New() (*Store, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	dir := filepath.Join(configDir, "taskscheduler")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	fp := filepath.Join(dir, "tasks.json")
	s := &Store{filePath: fp}

	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load tasks: %w", err)
	}

	return s, nil
}

// load reads tasks from disk.
func (s *Store) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.tasks)
}

// save writes tasks to disk atomically.
func (s *Store) save() error {
	data, err := json.MarshalIndent(s.tasks, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, s.filePath)
}

// Add inserts a new task and persists it.
func (s *Store) Add(task model.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	task.CreatedAt = time.Now()
	s.tasks = append(s.tasks, task)
	return s.save()
}

// All() returns a copy of all tasks (read-lock only for concurrency).
func (s *Store) All() []model.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Task, len(s.tasks))
	copy(out, s.tasks)
	return out
}

// Update replaces a task by ID.
func (s *Store) Update(task model.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.tasks {
		if t.ID == task.ID {
			s.tasks[i] = task
			return s.save()
		}
	}
	return fmt.Errorf("task %s not found", task.ID)
}

// Delete removes a task by ID.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.tasks {
		if t.ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("task %s not found", id)
}

// Get retrieves a task by ID.
func (s *Store) Get(id string) (model.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tasks {
		if t.ID == id {
			return t, true
		}
	}
	return model.Task{}, false
}

// Count returns the number of tasks.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tasks)
}
