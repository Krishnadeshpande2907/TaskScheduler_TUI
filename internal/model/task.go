package model

import (
	"encoding/json"
	"time"
)

// Priority represents the urgency level of a task.
type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "Low"
	case PriorityMedium:
		return "Medium"
	case PriorityHigh:
		return "High"
	case PriorityCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// PriorityColor returns an ANSI-friendly color hint for the priority.
func (p Priority) Color() string {
	switch p {
	case PriorityLow:
		return "#6c757d" // grey
	case PriorityMedium:
		return "#0d6efd" // blue
	case PriorityHigh:
		return "#fd7e14" // orange
	case PriorityCritical:
		return "#dc3545" // red
	default:
		return "#adb5bd"
	}
}

// Status represents the lifecycle state of a task.
type Status int

const (
	StatusPending Status = iota
	StatusInProgress
	StatusDone
	StatusOverdue
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "Pending"
	case StatusInProgress:
		return "In Progress"
	case StatusDone:
		return "Done"
	case StatusOverdue:
		return "Overdue"
	default:
		return "Unknown"
	}
}

// ReminderConfig holds the configuration for repeating reminders.
type ReminderConfig struct {
	Enabled      bool          `json:"enabled"`
	Repeating    bool          `json:"repeating"`
	Count        int           `json:"count"`         // number of reminder repetitions
	IntervalMins int           `json:"interval_mins"` // minutes between repeating reminders
	Interval     time.Duration `json:"-"`             // computed from IntervalMins
}

// Task is the core domain entity representing a scheduled task.
type Task struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Details     string          `json:"details,omitempty"`
	Priority    Priority        `json:"priority"`
	Status      Status          `json:"status"`
	ScheduledAt time.Time       `json:"scheduled_at"`        // when the task should be started
	Deadline    *time.Time      `json:"deadline,omitempty"`   // optional deadline
	StartedAt   *time.Time      `json:"started_at,omitempty"` // when the user started
	FinishedAt  *time.Time      `json:"finished_at,omitempty"`
	Reminder    ReminderConfig  `json:"reminder"`
	CreatedAt   time.Time       `json:"created_at"`
}

// Duration returns the time spent on the task, or zero if not started/finished.
func (t *Task) Duration() time.Duration {
	if t.StartedAt == nil || t.FinishedAt == nil {
		return 0
	}
	return t.FinishedAt.Sub(*t.StartedAt)
}

// IsOverdue checks whether the task has exceeded its deadline.
func (t *Task) IsOverdue() bool {
	if t.Deadline == nil {
		return false
	}
	return time.Now().After(*t.Deadline) && t.Status != StatusDone
}

// TimeUntilScheduled returns the duration until the task is scheduled.
func (t *Task) TimeUntilScheduled() time.Duration {
	return time.Until(t.ScheduledAt)
}

// TimeUntilDeadline returns the duration until the deadline, or 0 if none.
func (t *Task) TimeUntilDeadline() time.Duration {
	if t.Deadline == nil {
		return 0
	}
	return time.Until(*t.Deadline)
}

// MarshalJSON implements custom JSON marshaling for Task.
func (t Task) MarshalJSON() ([]byte, error) {
	type Alias Task
	return json.Marshal(&struct {
		Alias
	}{
		Alias: Alias(t),
	})
}
