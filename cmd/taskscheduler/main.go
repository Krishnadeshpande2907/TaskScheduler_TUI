package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Krishnadeshpande2907/TaskSchedulerTUI/internal/scheduler"
	"github.com/Krishnadeshpande2907/TaskSchedulerTUI/internal/store"
	"github.com/Krishnadeshpande2907/TaskSchedulerTUI/internal/tui"
)

func main() {
	// Initialize persistent store.
	taskStore, err := store.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing store: %v\n", err)
		os.Exit(1)
	}

	// Notification channel (buffered for non-blocking sends from scheduler).
	notifyCh := make(chan scheduler.Notification, 32)

	// Initialize scheduler and re-schedule persisted tasks.
	sched := scheduler.New(notifyCh)
	for _, t := range taskStore.All() {
		sched.Schedule(t)
	}

	// Build and run the TUI.
	m := tui.NewModel(taskStore, sched, notifyCh)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}

	sched.StopAll()
}
