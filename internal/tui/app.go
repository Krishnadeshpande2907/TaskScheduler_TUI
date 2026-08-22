package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Krishnadeshpande2907/TaskSchedulerTUI/internal/model"
	"github.com/Krishnadeshpande2907/TaskSchedulerTUI/internal/scheduler"
	"github.com/Krishnadeshpande2907/TaskSchedulerTUI/internal/store"
	"github.com/Krishnadeshpande2907/TaskSchedulerTUI/internal/tui/theme"
)

// ── Views ──────────────────────────────────────────────────────────
type viewMode int

const (
	viewList viewMode = iota
	viewAdd
	viewDetail
)

// ── Form fields ────────────────────────────────────────────────────
const (
	fieldName = iota
	fieldDetails
	fieldHour
	fieldMinute
	fieldDay
	fieldMonth
	fieldYear
	fieldPriority
	fieldDeadlineYN
	fieldDeadlineHour
	fieldDeadlineMin
	fieldDeadlineDay
	fieldDeadlineMonth
	fieldDeadlineYear
	fieldReminderYN
	fieldReminderRepeat
	fieldReminderCount
	fieldReminderInterval
	fieldCount
)

// ── Messages ───────────────────────────────────────────────────────
type tickMsg time.Time
type notificationMsg scheduler.Notification

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func waitForNotification(ch chan scheduler.Notification) tea.Cmd {
	return func() tea.Msg {
		n := <-ch
		return notificationMsg(n)
	}
}

// ── Model ──────────────────────────────────────────────────────────
type Model struct {
	store     *store.Store
	scheduler *scheduler.Scheduler
	notifyCh  chan scheduler.Notification

	view       viewMode
	tasks      []model.Task
	cursor     int
	width      int
	height     int
	now        time.Time
	notification string
	notifyTimer  int

	// Form
	inputs     []textinput.Model
	formCursor int

	// Detail
	detailTask *model.Task

	quitting bool
}

func NewModel(s *store.Store, sched *scheduler.Scheduler, ch chan scheduler.Notification) Model {
	inputs := make([]textinput.Model, fieldCount)
	placeholders := [fieldCount]string{
		"Task name",
		"Optional description",
		"Hour (0-23)",
		"Minute (0-59)",
		fmt.Sprintf("Day (default: %d)", time.Now().Day()),
		fmt.Sprintf("Month (default: %d)", int(time.Now().Month())),
		fmt.Sprintf("Year (default: %d)", time.Now().Year()),
		"0=Low 1=Med 2=High 3=Critical",
		"Deadline? (y/n)",
		"Deadline Hour",
		"Deadline Minute",
		"Deadline Day",
		"Deadline Month",
		"Deadline Year",
		"Reminder? (y/n)",
		"Repeat? (y/n)",
		"How many times?",
		"Interval (minutes)",
	}
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = placeholders[i]
		inputs[i].CharLimit = 64
		inputs[i].Width = 30
	}
	inputs[0].Focus()

	m := Model{
		store:     s,
		scheduler: sched,
		notifyCh:  ch,
		view:      viewList,
		inputs:    inputs,
		now:       time.Now(),
	}
	m.refreshTasks()
	return m
}

func (m *Model) refreshTasks() {
	m.tasks = m.store.All()
	// Update overdue status
	for i := range m.tasks {
		if m.tasks[i].IsOverdue() && m.tasks[i].Status != model.StatusDone {
			m.tasks[i].Status = model.StatusOverdue
			_ = m.store.Update(m.tasks[i])
		}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), waitForNotification(m.notifyCh), textinput.Blink)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		m.now = time.Time(msg)
		if m.notifyTimer > 0 {
			m.notifyTimer--
			if m.notifyTimer == 0 {
				m.notification = ""
			}
		}
		m.refreshTasks()
		return m, tickCmd()

	case notificationMsg:
		n := scheduler.Notification(msg)
		m.notification = fmt.Sprintf("%s — %s", n.TaskName, n.Message)
		m.notifyTimer = 8
		m.refreshTasks()
		return m, waitForNotification(m.notifyCh)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		switch m.view {
		case viewList:
			return m.updateList(msg)
		case viewAdd:
			return m.updateForm(msg)
		case viewDetail:
			return m.updateDetail(msg)
		}
	}
	return m, nil
}

// ── List view update ───────────────────────────────────────────────
func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.tasks)-1 {
			m.cursor++
		}
	case "a":
		m.view = viewAdd
		m.formCursor = 0
		for i := range m.inputs {
			m.inputs[i].Reset()
		}
		m.inputs[0].Focus()
		return m, textinput.Blink
	case "enter":
		if len(m.tasks) > 0 {
			t := m.tasks[m.cursor]
			m.detailTask = &t
			m.view = viewDetail
		}
	case "s":
		if len(m.tasks) > 0 {
			t := m.tasks[m.cursor]
			if t.Status == model.StatusPending || t.Status == model.StatusOverdue {
				now := time.Now()
				t.StartedAt = &now
				t.Status = model.StatusInProgress
				_ = m.store.Update(t)
				m.refreshTasks()
			}
		}
	case "d":
		if len(m.tasks) > 0 {
			t := m.tasks[m.cursor]
			if t.Status == model.StatusInProgress {
				now := time.Now()
				t.FinishedAt = &now
				t.Status = model.StatusDone
				_ = m.store.Update(t)
				m.scheduler.Cancel(t.ID)
				m.refreshTasks()
			}
		}
	case "x":
		if len(m.tasks) > 0 {
			_ = m.store.Delete(m.tasks[m.cursor].ID)
			m.scheduler.Cancel(m.tasks[m.cursor].ID)
			m.refreshTasks()
			if m.cursor >= len(m.tasks) && m.cursor > 0 {
				m.cursor--
			}
		}
	}
	return m, nil
}

// ── Form view update ───────────────────────────────────────────────
func (m Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view = viewList
		return m, nil
	case "tab", "down":
		m.inputs[m.formCursor].Blur()
		m.formCursor = (m.formCursor + 1) % fieldCount
		m.inputs[m.formCursor].Focus()
		return m, textinput.Blink
	case "shift+tab", "up":
		m.inputs[m.formCursor].Blur()
		m.formCursor = (m.formCursor - 1 + fieldCount) % fieldCount
		m.inputs[m.formCursor].Focus()
		return m, textinput.Blink
	case "ctrl+s":
		task, err := m.buildTaskFromForm()
		if err != nil {
			m.notification = "❌ " + err.Error()
			m.notifyTimer = 5
			return m, nil
		}
		if err := m.store.Add(task); err != nil {
			m.notification = "❌ Save failed: " + err.Error()
			m.notifyTimer = 5
			return m, nil
		}
		m.scheduler.Schedule(task)
		m.notification = "✅ Task created: " + task.Name
		m.notifyTimer = 5
		m.view = viewList
		m.refreshTasks()
		return m, nil
	}

	var cmd tea.Cmd
	m.inputs[m.formCursor], cmd = m.inputs[m.formCursor].Update(msg)
	return m, cmd
}

func (m Model) buildTaskFromForm() (model.Task, error) {
	name := strings.TrimSpace(m.inputs[fieldName].Value())
	if name == "" {
		return model.Task{}, fmt.Errorf("task name is required")
	}

	now := time.Now()
	hour := intOrDefault(m.inputs[fieldHour].Value(), now.Hour())
	min := intOrDefault(m.inputs[fieldMinute].Value(), now.Minute())
	day := intOrDefault(m.inputs[fieldDay].Value(), now.Day())
	month := intOrDefault(m.inputs[fieldMonth].Value(), int(now.Month()))
	year := intOrDefault(m.inputs[fieldYear].Value(), now.Year())
	prio := intOrDefault(m.inputs[fieldPriority].Value(), 0)

	scheduled := time.Date(year, time.Month(month), day, hour, min, 0, 0, time.Local)

	task := model.Task{
		ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Name:        name,
		Details:     strings.TrimSpace(m.inputs[fieldDetails].Value()),
		Priority:    model.Priority(prio),
		Status:      model.StatusPending,
		ScheduledAt: scheduled,
		CreatedAt:   now,
	}

	if strings.EqualFold(strings.TrimSpace(m.inputs[fieldDeadlineYN].Value()), "y") {
		dh := intOrDefault(m.inputs[fieldDeadlineHour].Value(), hour+1)
		dm := intOrDefault(m.inputs[fieldDeadlineMin].Value(), min)
		dd := intOrDefault(m.inputs[fieldDeadlineDay].Value(), day)
		dmo := intOrDefault(m.inputs[fieldDeadlineMonth].Value(), month)
		dy := intOrDefault(m.inputs[fieldDeadlineYear].Value(), year)
		dl := time.Date(dy, time.Month(dmo), dd, dh, dm, 0, 0, time.Local)
		task.Deadline = &dl
	}

	if strings.EqualFold(strings.TrimSpace(m.inputs[fieldReminderYN].Value()), "y") {
		task.Reminder.Enabled = true
		if strings.EqualFold(strings.TrimSpace(m.inputs[fieldReminderRepeat].Value()), "y") {
			task.Reminder.Repeating = true
			task.Reminder.Count = intOrDefault(m.inputs[fieldReminderCount].Value(), 3)
			task.Reminder.IntervalMins = intOrDefault(m.inputs[fieldReminderInterval].Value(), 5)
		}
	}

	return task, nil
}

// ── Detail view update ─────────────────────────────────────────────
func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace":
		m.view = viewList
	case "s":
		if m.detailTask != nil && (m.detailTask.Status == model.StatusPending || m.detailTask.Status == model.StatusOverdue) {
			now := time.Now()
			m.detailTask.StartedAt = &now
			m.detailTask.Status = model.StatusInProgress
			_ = m.store.Update(*m.detailTask)
			m.refreshTasks()
		}
	case "d":
		if m.detailTask != nil && m.detailTask.Status == model.StatusInProgress {
			now := time.Now()
			m.detailTask.FinishedAt = &now
			m.detailTask.Status = model.StatusDone
			_ = m.store.Update(*m.detailTask)
			m.scheduler.Cancel(m.detailTask.ID)
			m.refreshTasks()
		}
	}
	return m, nil
}

// ── View rendering ─────────────────────────────────────────────────
func (m Model) View() string {
	if m.quitting {
		return lipgloss.NewStyle().Foreground(theme.ColorPrimary).Render("\n  👋 Goodbye! Tasks saved.\n\n")
	}

	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Main content
	switch m.view {
	case viewList:
		b.WriteString(m.renderList())
	case viewAdd:
		b.WriteString(m.renderForm())
	case viewDetail:
		b.WriteString(m.renderDetail())
	}

	// Notification
	if m.notification != "" {
		b.WriteString("\n")
		b.WriteString(theme.NotificationPanel.Render(m.notification))
	}

	// Help bar
	b.WriteString("\n")
	b.WriteString(m.renderHelp())

	return b.String()
}

func (m Model) renderHeader() string {
	clock := theme.TaskTime.Render(m.now.Format("15:04:05"))
	date := theme.TaskTime.Render(m.now.Format("Mon, 02 Jan 2006"))

	title := theme.Title.Render("⚡ Task Scheduler")
	stats := fmt.Sprintf("  %s │ %s │ %d tasks",
		date, clock, len(m.tasks))
	statsStyled := lipgloss.NewStyle().Foreground(theme.ColorMuted).Render(stats)

	// Tabs
	tabs := ""
	tabNames := []string{"Tasks", "Add New"}
	for i, name := range tabNames {
		active := (i == 0 && m.view == viewList) || (i == 1 && m.view == viewAdd)
		if active {
			tabs += theme.TabActive.Render(name)
		} else {
			tabs += theme.TabInactive.Render(name)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Center, title, statsStyled),
		lipgloss.NewStyle().Foreground(theme.ColorBorder).Render(strings.Repeat("─", min(m.width, 80))),
		tabs,
	)
}

func (m Model) renderList() string {
	if len(m.tasks) == 0 {
		empty := lipgloss.NewStyle().
			Foreground(theme.ColorMuted).
			Padding(2, 4).
			Render("No tasks yet. Press 'a' to add your first task!")
		return theme.Panel.Render(empty)
	}

	var rows []string
	for i, t := range m.tasks {
		rows = append(rows, m.renderTaskRow(i, t))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return theme.Panel.Width(min(m.width-4, 78)).Render(content)
}

func (m Model) renderTaskRow(idx int, t model.Task) string {
	cursor := "  "
	if idx == m.cursor {
		cursor = "▸ "
	}

	// Status icon
	var statusIcon string
	var statusStyle lipgloss.Style
	switch t.Status {
	case model.StatusPending:
		statusIcon = "○"
		statusStyle = theme.StatusPending
	case model.StatusInProgress:
		statusIcon = "◉"
		statusStyle = theme.StatusInProgress
	case model.StatusDone:
		statusIcon = "✓"
		statusStyle = theme.StatusDone
	case model.StatusOverdue:
		statusIcon = "✗"
		statusStyle = theme.StatusOverdue
	}

	// Priority
	prioStyle := theme.PriorityBadge(t.Priority.Color())
	prioLabel := prioStyle.Render(fmt.Sprintf("[%s]", t.Priority.String()))

	// Time info
	timeStr := theme.TaskTime.Render(t.ScheduledAt.Format("15:04"))
	deadlineStr := ""
	if t.Deadline != nil {
		remaining := time.Until(*t.Deadline)
		if remaining > 0 {
			deadlineStr = lipgloss.NewStyle().Foreground(theme.ColorWarning).
				Render(fmt.Sprintf(" ⏱ %s left", formatDuration(remaining)))
		} else {
			deadlineStr = lipgloss.NewStyle().Foreground(theme.ColorDanger).
				Render(" ⏱ overdue")
		}
	}

	// Duration if done
	durStr := ""
	if t.Status == model.StatusDone && t.StartedAt != nil && t.FinishedAt != nil {
		durStr = lipgloss.NewStyle().Foreground(theme.ColorSuccess).
			Render(fmt.Sprintf(" took %s", formatDuration(t.Duration())))
	}

	line := fmt.Sprintf("%s%s %s %s %s%s%s",
		cursor,
		statusStyle.Render(statusIcon),
		theme.TaskName.Render(t.Name),
		prioLabel,
		timeStr,
		deadlineStr,
		durStr,
	)

	if idx == m.cursor {
		return theme.TaskItemSelected.Render(line)
	}
	return theme.TaskItem.Render(line)
}

func (m Model) renderForm() string {
	var b strings.Builder

	title := theme.Subtitle.Render("📝 New Task")
	b.WriteString(title)
	b.WriteString("\n\n")

	labels := [fieldCount]string{
		"Task Name*:",
		"Details:",
		"Hour (0-23)*:",
		"Minute (0-59)*:",
		"Day:",
		"Month:",
		"Year:",
		"Priority:",
		"Set Deadline?:",
		"Deadline Hour:",
		"Deadline Min:",
		"Deadline Day:",
		"Deadline Month:",
		"Deadline Year:",
		"Set Reminder?:",
		"Repeat?:",
		"Repeat Count:",
		"Interval (min):",
	}

	sections := []struct {
		name   string
		fields []int
	}{
		{"Task Info", []int{fieldName, fieldDetails}},
		{"Schedule", []int{fieldHour, fieldMinute, fieldDay, fieldMonth, fieldYear, fieldPriority}},
		{"Deadline", []int{fieldDeadlineYN, fieldDeadlineHour, fieldDeadlineMin, fieldDeadlineDay, fieldDeadlineMonth, fieldDeadlineYear}},
		{"Reminders", []int{fieldReminderYN, fieldReminderRepeat, fieldReminderCount, fieldReminderInterval}},
	}

	for _, sec := range sections {
		b.WriteString(lipgloss.NewStyle().
			Foreground(theme.ColorAccent).Bold(true).
			Render("── " + sec.name + " "))
		b.WriteString(lipgloss.NewStyle().Foreground(theme.ColorBorder).
			Render(strings.Repeat("─", 40)))
		b.WriteString("\n")

		for _, fi := range sec.fields {
			indicator := "  "
			if fi == m.formCursor {
				indicator = "▸ "
			}
			label := theme.FormLabel.Render(labels[fi])
			b.WriteString(fmt.Sprintf("%s%s %s\n", indicator, label, m.inputs[fi].View()))
		}
		b.WriteString("\n")
	}

	return theme.Panel.Width(min(m.width-4, 78)).Render(b.String())
}

func (m Model) renderDetail() string {
	if m.detailTask == nil {
		return ""
	}
	t := m.detailTask

	// Re-fetch latest state
	if latest, ok := m.store.Get(t.ID); ok {
		t = &latest
		m.detailTask = t
	}

	var b strings.Builder
	b.WriteString(theme.Subtitle.Render("📋 Task Details"))
	b.WriteString("\n\n")

	row := func(label, value string) {
		b.WriteString(fmt.Sprintf("  %s %s\n",
			theme.FormLabel.Render(label+":"),
			lipgloss.NewStyle().Foreground(theme.ColorText).Render(value),
		))
	}

	row("Name", t.Name)
	if t.Details != "" {
		row("Details", t.Details)
	}
	row("Priority", t.Priority.String())
	row("Status", t.Status.String())
	row("Scheduled", t.ScheduledAt.Format("2006-01-02 15:04"))

	if t.Deadline != nil {
		row("Deadline", t.Deadline.Format("2006-01-02 15:04"))
		remaining := time.Until(*t.Deadline)
		if remaining > 0 {
			row("Time Left", formatDuration(remaining))
		} else {
			row("Overdue By", formatDuration(-remaining))
		}
	}

	if t.StartedAt != nil {
		row("Started At", t.StartedAt.Format("15:04:05"))
	}
	if t.FinishedAt != nil {
		row("Finished At", t.FinishedAt.Format("15:04:05"))
		row("Duration", formatDuration(t.Duration()))
	}

	if t.Reminder.Enabled {
		if t.Reminder.Repeating {
			row("Reminders", fmt.Sprintf("%d times every %d min", t.Reminder.Count, t.Reminder.IntervalMins))
		} else {
			row("Reminder", "At scheduled time")
		}
	}

	row("Created", t.CreatedAt.Format("2006-01-02 15:04"))

	return theme.PanelActive.Width(min(m.width-4, 78)).Render(b.String())
}

func (m Model) renderHelp() string {
	var keys []string
	switch m.view {
	case viewList:
		keys = []string{
			"↑/↓", "navigate",
			"a", "add task",
			"enter", "details",
			"s", "start",
			"d", "done",
			"x", "delete",
			"q", "quit",
		}
	case viewAdd:
		keys = []string{
			"tab/↑↓", "fields",
			"ctrl+s", "save",
			"esc", "cancel",
		}
	case viewDetail:
		keys = []string{
			"s", "start",
			"d", "done",
			"esc", "back",
		}
	}

	var parts []string
	for i := 0; i < len(keys)-1; i += 2 {
		parts = append(parts, fmt.Sprintf("%s %s",
			theme.HelpKey.Render(keys[i]),
			theme.HelpDesc.Render(keys[i+1]),
		))
	}
	return theme.HelpBar.Render(strings.Join(parts, "  │  "))
}

// ── Helpers ────────────────────────────────────────────────────────
func intOrDefault(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
