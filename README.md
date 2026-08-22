# ⚡ Task Scheduler TUI

A beautiful terminal-based Task Scheduler built in Go, using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework.

Rewritten from the [original Java version](https://github.com/Krishnadeshpande2907/Task_Scheduler) with significant performance and UX improvements.

## Features

- **Rich TUI** — Full-screen interactive terminal UI with colors, borders, and keyboard navigation
- **Task Lifecycle** — Create → Start → Complete with time tracking
- **Priority Levels** — Low, Medium, High, Critical with color-coded badges
- **Deadline Monitoring** — Background goroutines warn you 5 minutes before and when a deadline passes
- **Repeating Reminders** — Set recurring reminders at configurable intervals
- **Persistent Storage** — Tasks are saved to `~/.config/taskscheduler/tasks.json` and survive restarts
- **Live Clock** — Real-time clock display and countdown timers

## Performance Improvements over Java

| Aspect | Java (Original) | Go (TUI) |
|--------|-----------------|----------|
| Startup | ~500ms (JVM cold start) | ~5ms (native binary) |
| Memory | ~50MB (JVM heap) | ~3MB (no runtime overhead) |
| Scheduling | Busy-wait `while` loops, `ScheduledExecutorService` | `time.AfterFunc` — zero CPU when idle |
| Concurrency | Thread pool + `synchronized` | Goroutines + `sync.RWMutex` |
| Persistence | None (in-memory only) | JSON file with atomic writes |
| Binary | Requires JRE installed | Single static binary |

## Installation

```bash
go install github.com/Krishnadeshpande2907/TaskScheduler_TUI/cmd/taskscheduler@latest
```

Or build from source:

```bash
git clone https://github.com/Krishnadeshpande2907/TaskScheduler_TUI.git
cd TaskScheduler_TUI
go build -o taskscheduler ./cmd/taskscheduler/
./taskscheduler
```

## Keybindings

### Task List
| Key | Action |
|-----|--------|
| `↑`/`k` | Move up |
| `↓`/`j` | Move down |
| `a` | Add new task |
| `Enter` | View task details |
| `s` | Start selected task |
| `d` | Mark task as done |
| `x` | Delete task |
| `q` | Quit |

### Add Task Form
| Key | Action |
|-----|--------|
| `Tab`/`↓` | Next field |
| `Shift+Tab`/`↑` | Previous field |
| `Ctrl+S` | Save task |
| `Esc` | Cancel |

### Task Details
| Key | Action |
|-----|--------|
| `s` | Start task |
| `d` | Mark done |
| `Esc` | Back to list |

## Technologies

- **Go** — Fast, compiled, zero-dependency runtime
- **Bubble Tea** — Elm-architecture TUI framework
- **Lip Gloss** — Styling and layout
- **Bubbles** — Text input components
