// Package theme defines the visual styling for the TUI using Lip Gloss.
package theme

import "github.com/charmbracelet/lipgloss"

// ── Color palette ──────────────────────────────────────────────────
var (
	ColorPrimary    = lipgloss.Color("#7C3AED") // vibrant purple
	ColorSecondary  = lipgloss.Color("#06B6D4") // cyan
	ColorAccent     = lipgloss.Color("#F59E0B") // amber
	ColorSuccess    = lipgloss.Color("#10B981") // emerald
	ColorDanger     = lipgloss.Color("#EF4444") // red
	ColorWarning    = lipgloss.Color("#F97316") // orange
	ColorMuted      = lipgloss.Color("#6B7280") // grey
	ColorText       = lipgloss.Color("#E5E7EB") // light grey
	ColorSubtle     = lipgloss.Color("#4B5563") // dark grey
	ColorBg         = lipgloss.Color("#0F172A") // dark navy
	ColorBgAlt      = lipgloss.Color("#1E293B") // slightly lighter navy
	ColorBorder     = lipgloss.Color("#334155") // border grey
	ColorHighlight  = lipgloss.Color("#A78BFA") // light purple
)

// ── Base styles ────────────────────────────────────────────────────

// AppContainer is the outermost wrapper.
var AppContainer = lipgloss.NewStyle().
	Background(ColorBg)

// Title renders the application header.
var Title = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorPrimary).
	PaddingLeft(2).
	PaddingRight(2)

// Subtitle for secondary headings.
var Subtitle = lipgloss.NewStyle().
	Foreground(ColorSecondary).
	Bold(true)

// ── Navigation / Tabs ──────────────────────────────────────────────

// TabActive is an active nav tab.
var TabActive = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FFFFFF")).
	Background(ColorPrimary).
	Padding(0, 2).
	MarginRight(1)

// TabInactive is an unselected tab.
var TabInactive = lipgloss.NewStyle().
	Foreground(ColorMuted).
	Padding(0, 2).
	MarginRight(1)

// ── Task list ──────────────────────────────────────────────────────

// TaskItem is the base style for a task row.
var TaskItem = lipgloss.NewStyle().
	PaddingLeft(2).
	PaddingRight(2).
	MarginBottom(0)

// TaskItemSelected highlights the currently selected task.
var TaskItemSelected = TaskItem.
	Bold(true).
	Foreground(lipgloss.Color("#FFFFFF")).
	Background(ColorBgAlt).
	BorderLeft(true).
	BorderStyle(lipgloss.ThickBorder()).
	BorderForeground(ColorPrimary)

// TaskName styles the task name text.
var TaskName = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorText)

// TaskDetail styles the detail/description text.
var TaskDetail = lipgloss.NewStyle().
	Foreground(ColorMuted).
	Italic(true)

// TaskTime styles time labels.
var TaskTime = lipgloss.NewStyle().
	Foreground(ColorSecondary)

// ── Status badges ──────────────────────────────────────────────────

// StatusPending badge.
var StatusPending = lipgloss.NewStyle().
	Foreground(ColorAccent).
	Bold(true)

// StatusInProgress badge.
var StatusInProgress = lipgloss.NewStyle().
	Foreground(ColorSecondary).
	Bold(true)

// StatusDone badge.
var StatusDone = lipgloss.NewStyle().
	Foreground(ColorSuccess).
	Bold(true)

// StatusOverdue badge.
var StatusOverdue = lipgloss.NewStyle().
	Foreground(ColorDanger).
	Bold(true)

// ── Priority badges ────────────────────────────────────────────────

func PriorityBadge(color string) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true)
}

// ── Panels / Cards ─────────────────────────────────────────────────

// Panel is a bordered content area.
var Panel = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(ColorBorder).
	Padding(1, 2)

// PanelActive has a highlighted border for the focused panel.
var PanelActive = Panel.
	BorderForeground(ColorPrimary)

// NotificationPanel for toast-like notifications.
var NotificationPanel = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(ColorWarning).
	Foreground(ColorAccent).
	Padding(0, 2).
	MarginTop(1)

// ── Form elements ──────────────────────────────────────────────────

// FormLabel for input labels.
var FormLabel = lipgloss.NewStyle().
	Foreground(ColorHighlight).
	Bold(true).
	MarginRight(1)

// FormHelp for hint text below inputs.
var FormHelp = lipgloss.NewStyle().
	Foreground(ColorMuted).
	Italic(true)

// ── Help bar ───────────────────────────────────────────────────────

// HelpBar at the bottom of the screen.
var HelpBar = lipgloss.NewStyle().
	Foreground(ColorSubtle).
	Padding(0, 2)

// HelpKey highlights a key binding.
var HelpKey = lipgloss.NewStyle().
	Foreground(ColorHighlight).
	Bold(true)

// HelpDesc describes a key binding.
var HelpDesc = lipgloss.NewStyle().
	Foreground(ColorMuted)

// ── Utility ────────────────────────────────────────────────────────

// Divider returns a horizontal separator line.
func Divider(width int) string {
	return lipgloss.NewStyle().
		Foreground(ColorBorder).
		Render(lipgloss.PlaceHorizontal(width, lipgloss.Left, "─────────────────────────────────────────────────────────"))
}

// Logo returns the ASCII art logo for the splash.
func Logo() string {
	logo := `
  ╔════════════════════════════════════════════╗
  ║                                            ║
  ║     ████████  █████  ███████ ██   ██       ║
  ║        ██    ██   ██ ██      ██  ██        ║
  ║        ██    ███████ ███████ █████         ║
  ║        ██    ██   ██      ██ ██  ██        ║
  ║        ██    ██   ██ ███████ ██   ██       ║
  ║                                            ║
  ║   ███████  ██████ ██   ██ ███████ ██████   ║
  ║   ██      ██      ██   ██ ██      ██   ██  ║
  ║   ███████ ██      ███████ █████   ██   ██  ║
  ║        ██ ██      ██   ██ ██      ██   ██  ║
  ║   ███████  ██████ ██   ██ ███████ ██████   ║
  ║                                            ║
  ╚════════════════════════════════════════════╝`

	return lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		Render(logo)
}
