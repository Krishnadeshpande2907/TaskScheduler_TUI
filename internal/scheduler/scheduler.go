// Package scheduler manages background goroutines for reminders and
// deadline monitoring. Uses Go's efficient time.Timer/time.Ticker
// instead of Java's busy-wait loops and ScheduledExecutorService.
package scheduler

import (
	"sync"
	"time"

	"github.com/Krishnadeshpande2907/TaskSchedulerTUI/internal/model"
)

// Notification represents a message from the scheduler to the TUI.
type Notification struct {
	TaskID  string
	TaskName string
	Message string
	Type    NotificationType
	Time    time.Time
}

// NotificationType categorizes scheduler notifications.
type NotificationType int

const (
	NotifyReminder NotificationType = iota
	NotifyDeadlineWarning
	NotifyDeadlinePassed
	NotifyTaskTime
)

// Scheduler manages timers for tasks.
type Scheduler struct {
	mu       sync.Mutex
	timers   map[string][]*time.Timer
	notifyCh chan Notification
}

// New creates a new Scheduler that sends notifications on the given channel.
func New(notifyCh chan Notification) *Scheduler {
	return &Scheduler{
		timers:   make(map[string][]*time.Timer),
		notifyCh: notifyCh,
	}
}

// Schedule sets up all timers for a task: scheduled time alert,
// deadline monitoring, and repeating reminders.
func (s *Scheduler) Schedule(task model.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel any existing timers for this task.
	s.cancelLocked(task.ID)

	var timers []*time.Timer

	// 1. Scheduled time notification.
	if d := time.Until(task.ScheduledAt); d > 0 {
		t := time.AfterFunc(d, func() {
			s.notifyCh <- Notification{
				TaskID:   task.ID,
				TaskName: task.Name,
				Message:  "⏰ It's time to start your task!",
				Type:     NotifyTaskTime,
				Time:     time.Now(),
			}
		})
		timers = append(timers, t)
	}

	// 2. Deadline warning (5 min before) and expiry notification.
	if task.Deadline != nil {
		warn := time.Until(task.Deadline.Add(-5 * time.Minute))
		if warn > 0 {
			t := time.AfterFunc(warn, func() {
				s.notifyCh <- Notification{
					TaskID:   task.ID,
					TaskName: task.Name,
					Message:  "⚠️  Deadline in 5 minutes!",
					Type:     NotifyDeadlineWarning,
					Time:     time.Now(),
				}
			})
			timers = append(timers, t)
		}

		expire := time.Until(*task.Deadline)
		if expire > 0 {
			t := time.AfterFunc(expire, func() {
				s.notifyCh <- Notification{
					TaskID:   task.ID,
					TaskName: task.Name,
					Message:  "🚨 Deadline has passed!",
					Type:     NotifyDeadlinePassed,
					Time:     time.Now(),
				}
			})
			timers = append(timers, t)
		}
	}

	// 3. Repeating reminders — uses efficient goroutine + time.Sleep.
	if task.Reminder.Enabled && task.Reminder.Repeating && task.Reminder.Count > 0 {
		interval := time.Duration(task.Reminder.IntervalMins) * time.Minute
		startDelay := time.Until(task.ScheduledAt)
		if startDelay < 0 {
			startDelay = 0
		}

		// A single goroutine handles all repeating reminders for this task.
		// Much lighter than Java's ScheduledExecutorService thread pool.
		reminderTimer := time.AfterFunc(startDelay, func() {
			for i := 0; i < task.Reminder.Count; i++ {
				if i > 0 {
					time.Sleep(interval)
				}
				s.notifyCh <- Notification{
					TaskID:   task.ID,
					TaskName: task.Name,
					Message:  "🔔 Reminder: Time to work on your task!",
					Type:     NotifyReminder,
					Time:     time.Now(),
				}
			}
		})
		timers = append(timers, reminderTimer)
	} else if task.Reminder.Enabled {
		// Single reminder at scheduled time (already handled by #1 above,
		// but we add an explicit reminder notification too).
		d := time.Until(task.ScheduledAt)
		if d > 0 {
			t := time.AfterFunc(d, func() {
				s.notifyCh <- Notification{
					TaskID:   task.ID,
					TaskName: task.Name,
					Message:  "🔔 Reminder: Your task is due now!",
					Type:     NotifyReminder,
					Time:     time.Now(),
				}
			})
			timers = append(timers, t)
		}
	}

	s.timers[task.ID] = timers
}

// Cancel stops all timers for a task.
func (s *Scheduler) Cancel(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelLocked(taskID)
}

func (s *Scheduler) cancelLocked(taskID string) {
	if timers, ok := s.timers[taskID]; ok {
		for _, t := range timers {
			t.Stop()
		}
		delete(s.timers, taskID)
	}
}

// StopAll cancels every active timer.
func (s *Scheduler) StopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id := range s.timers {
		s.cancelLocked(id)
	}
}
