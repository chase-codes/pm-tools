package models

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chase/pm-tools/issue-monitor/internal/aksmonitor/services"
	"github.com/google/go-github/v58/github"
)

type viewMode int

const (
	viewModeList viewMode = iota
	viewModeTable
	viewModeDetail
	viewModeComments
)

type GitHubIssuesModel struct {
	services          *services.Services
	table             table.Model
	viewport          viewport.Model
	previewPane       viewport.Model
	searchInput       textinput.Model
	filterInput       textinput.Model
	spinner           spinner.Model
	selected          *services.IssueWithRepo
	selectedIndex     int
	comments          []*github.IssueComment
	loading           bool
	loadingComments   bool
	error             string
	showFilters       bool
	showPreview       bool
	activeQuickFilter int
	currentView       viewMode
	issues            []services.IssueWithRepo
	filteredIssues    []services.IssueWithRepo
	currentColumns    []table.Column // Track current column configuration
	width             int
	height            int
}

var (
	// Enhanced color scheme with more semantic colors
	primaryColor   = lipgloss.Color("#00D7FF") // Bright cyan
	secondaryColor = lipgloss.Color("#7C3AED") // Purple
	successColor   = lipgloss.Color("#10B981") // Green
	warningColor   = lipgloss.Color("#F59E0B") // Amber
	errorColor     = lipgloss.Color("#EF4444") // Red
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	accentColor    = lipgloss.Color("#F97316") // Orange
	bgColor        = lipgloss.Color("#1F2937") // Dark gray
	borderColor    = lipgloss.Color("#374151") // Medium gray

	// Enhanced styles with better visual hierarchy
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Background(bgColor).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			MarginBottom(1)

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#000000")).
				Background(primaryColor).
				Padding(0, 1)

	filterBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			MarginBottom(1).
			Background(bgColor)

	quickFilterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(accentColor).
				Padding(0, 2).
				MarginRight(1).
				BorderStyle(lipgloss.RoundedBorder())

	quickFilterInactiveStyle = lipgloss.NewStyle().
					Foreground(mutedColor).
					Background(borderColor).
					Padding(0, 2).
					MarginRight(1).
					BorderStyle(lipgloss.RoundedBorder())

	statusBarStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Background(bgColor).
			Padding(0, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(borderColor)

	issueOpenStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	issueClosedStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Strikethrough(true)

	priorityHighStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	priorityMediumStyle = lipgloss.NewStyle().
				Foreground(warningColor)

	priorityLowStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(accentColor).
			Padding(0, 1).
			MarginRight(1)

	metaStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	detailHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				MarginBottom(1)

	detailContentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(1).
				Background(bgColor).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(borderColor)
)

type quickFilter struct {
	name  string
	query string
	key   string
}

var quickFilters = []quickFilter{
	{"My Issues", "assignee:@me", "1"},
	{"Open", "state:open", "2"},
	{"Recent", "updated:>7days", "3"},
	{"Bugs", "label:bug", "4"},
	{"Features", "label:feature", "5"},
	{"High Priority", "label:priority-high", "6"},
}

func NewGitHubIssuesModel(services *services.Services) *GitHubIssuesModel {
	// Create initial table columns - these will be adjusted based on terminal size
	initialColumns := []table.Column{
		{Title: "#", Width: 7},
		{Title: "Title", Width: 50},
		{Title: "State", Width: 9},
		{Title: "Assignee", Width: 15},
		{Title: "Labels", Width: 18},
		{Title: "Updated", Width: 10},
		{Title: "Repo", Width: 15},
	}

	// Create table with proper configuration
	t := table.New(
		table.WithColumns(initialColumns),
		table.WithFocused(true),
		table.WithHeight(20),
		table.WithWidth(120), // Add explicit width
	)

	// Configure table styles properly
	tableStyles := table.DefaultStyles()
	tableStyles.Header = tableStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(primaryColor).
		BorderBottom(true).
		Bold(true).
		Foreground(primaryColor)

	tableStyles.Selected = tableStyles.Selected.
		Bold(true).
		Foreground(lipgloss.Color("#000000")).
		Background(primaryColor)

	t.SetStyles(tableStyles)

	// Initialize viewport for detail view
	vp := viewport.New(100, 25)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1)

	// Initialize preview pane
	previewPane := viewport.New(50, 20)
	previewPane.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(secondaryColor).
		Padding(1)

	// Initialize search input
	searchInput := textinput.New()
	searchInput.Placeholder = "🔍 Search issues, labels, assignees..."
	searchInput.CharLimit = 100
	searchInput.Width = 50

	// Initialize filter input
	filterInput := textinput.New()
	filterInput.Placeholder = "🎯 Filter: state:open, label:bug, author:username..."
	filterInput.CharLimit = 100
	filterInput.Width = 50

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	return &GitHubIssuesModel{
		services:          services,
		table:             t,
		viewport:          vp,
		previewPane:       previewPane,
		searchInput:       searchInput,
		filterInput:       filterInput,
		spinner:           s,
		currentView:       viewModeTable,
		selectedIndex:     -1,
		activeQuickFilter: -1,
		showPreview:       true,
		currentColumns:    initialColumns, // Initialize with default columns
	}
}

func (m *GitHubIssuesModel) Init() tea.Cmd {
	// Start with reasonable defaults but ensure sizing works
	m.width = 150        // Default to wide screen to enable preview
	m.height = 40        // Reasonable height
	m.showPreview = true // Ensure preview is enabled

	// Set up initial table configuration
	m.updateSizes()

	return tea.Batch(
		m.loadIssues(),
		m.spinner.Tick,
	)
}

func (m *GitHubIssuesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()
		// Force table rows update after size change
		m.updateTableRows()
		return m, nil

	case tea.KeyMsg:
		// Check if any input is currently focused - if so, let inputs handle ALL keys
		inputFocused := m.searchInput.Focused() || m.filterInput.Focused()

		// Only handle special keys when no input is focused
		if !inputFocused {
			// Handle quick filter shortcuts (1-6)
			if m.currentView == viewModeTable {
				switch msg.String() {
				case "1", "2", "3", "4", "5", "6":
					filterIndex := int(msg.String()[0] - '1') // Convert '1'-'6' to 0-5
					if filterIndex < len(quickFilters) {
						if m.activeQuickFilter == filterIndex {
							// Toggle off if already active
							m.activeQuickFilter = -1
							m.filterInput.SetValue("")
						} else {
							// Apply quick filter
							m.activeQuickFilter = filterIndex
							m.filterInput.SetValue(quickFilters[filterIndex].query)
						}
						m.applyFilters()
						return m, nil
					}
				}
			}

			// Handle filter toggle
			if msg.String() == "f" && m.currentView == viewModeTable {
				if !m.showFilters {
					// Starting filter mode
					m.showFilters = true
					m.filterInput.Focus()
					return m, nil
				} else {
					// Exit filter mode only if no input is focused
					m.showFilters = false
					m.filterInput.Blur()
					m.searchInput.Blur()
					return m, nil
				}
			}

			// Handle search toggle
			if msg.String() == "s" && m.currentView == viewModeTable && !m.showFilters {
				m.searchInput.Focus()
				return m, nil
			}

			// Handle preview toggle
			if msg.String() == "p" && m.currentView == viewModeTable {
				m.showPreview = !m.showPreview
				return m, nil
			}
		}

		// Handle keys that should work regardless of input focus
		switch msg.String() {
		case "esc":
			if m.currentView == viewModeComments {
				m.currentView = viewModeDetail
				m.comments = nil
			} else if m.currentView == viewModeDetail {
				m.currentView = viewModeTable
				m.selected = nil
				m.selectedIndex = -1
			} else if m.showFilters {
				// Exit filter mode and blur all inputs
				m.showFilters = false
				m.searchInput.Blur()
				m.filterInput.Blur()
			} else if m.searchInput.Focused() {
				// Just blur search if it's focused
				m.searchInput.Blur()
			}
			return m, nil

		case "tab":
			if m.showFilters {
				if m.filterInput.Focused() {
					m.filterInput.Blur()
					m.searchInput.Focus()
				} else if m.searchInput.Focused() {
					m.searchInput.Blur()
					m.filterInput.Focus()
				} else {
					// No input focused, focus filter first
					m.filterInput.Focus()
				}
				return m, nil
			}
		}

		// Only handle these keys when NO input is focused
		if !inputFocused {
			switch msg.String() {
			case "enter":
				if m.currentView == viewModeTable && len(m.filteredIssues) > 0 {
					cursor := m.table.Cursor()
					if cursor < len(m.filteredIssues) {
						m.selected = &m.filteredIssues[cursor]
						m.selectedIndex = cursor
						m.currentView = viewModeDetail
						m.updateDetailView()
					}
				}

			case "r":
				return m, m.loadIssues()

			case "o":
				// Open selected issue in browser
				if m.selected != nil {
					return m, m.openInBrowser()
				}

			case "y":
				// Copy issue description to clipboard
				if m.selected != nil {
					return m, m.copyToClipboard()
				}

			case "c":
				// Show comments view
				if m.selected != nil {
					m.loadingComments = true
					return m, m.loadComments()
				}
			}
		}

	case issuesLoadedMsg:
		m.loading = false
		m.error = ""
		m.issues = msg.Issues
		m.applyFilters()

	case errorMsg:
		m.loading = false
		m.error = msg.Error

	case browserActionMsg:
		if msg.success {
			// Clear any previous error and show success message temporarily
			m.error = ""
			// You could add a success message field if desired
		} else {
			m.error = msg.message
		}

	case commentsActionMsg:
		if msg.success {
			// Clear any previous error
			m.error = ""
			// For now, we'll update the detail view to show the comment info
			if m.selected != nil {
				m.updateDetailViewWithCommentInfo(msg.message)
			}
		} else {
			m.error = msg.message
		}

	case commentsLoadedMsg:
		m.loadingComments = false
		m.comments = msg.comments
		m.currentView = viewModeComments
		m.updateCommentsView()

	case tea.MouseMsg:
		// Handle mouse events safely to prevent crashes
		switch msg.Type {
		case tea.MouseWheelUp, tea.MouseWheelDown:
			// Ignore mouse wheel events to prevent crashes
			return m, nil
		default:
			// Allow other mouse events to pass through
		}
	}

	// Update components based on current view
	if m.currentView == viewModeTable {
		// Handle input updates when they are focused
		if m.searchInput.Focused() {
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			cmds = append(cmds, cmd)
			// Apply search in real-time
			m.applyFilters()
		} else if m.filterInput.Focused() {
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)
			cmds = append(cmds, cmd)
			// Apply filters in real-time
			m.applyFilters()
		} else {
			// Only update table when no input is focused
			var cmd tea.Cmd
			oldCursor := m.table.Cursor()
			m.table, cmd = m.table.Update(msg)
			cmds = append(cmds, cmd)

			// Update preview if cursor moved and preview is enabled
			newCursor := m.table.Cursor()
			if m.showPreview && oldCursor != newCursor && len(m.filteredIssues) > 0 && newCursor < len(m.filteredIssues) {
				m.updatePreviewPane(newCursor)
			}
		}
	} else if m.currentView == viewModeDetail {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *GitHubIssuesModel) updateSizes() {
	// Use current dimensions or fallback to reasonable defaults
	width := m.width
	height := m.height

	if width <= 0 {
		width = 150 // Default wide width
	}
	if height <= 0 {
		height = 40 // Default height
	}

	// Calculate available space
	availableWidth := width - 4    // Account for borders/padding
	availableHeight := height - 15 // Account for main header (title + repo + tabs), filters, footer

	// Ensure minimum dimensions
	if availableWidth < 20 {
		availableWidth = 80 // Reasonable minimum
	}
	if availableHeight < 5 {
		availableHeight = 20 // Reasonable minimum
	}

	if m.showFilters {
		availableHeight -= 4 // Additional space for filter inputs
	}

	// Always enable preview if we have enough width
	if width > 120 {
		m.showPreview = true

		// Split view - adjust table columns for smaller space
		tableWidth := availableWidth * 60 / 100
		m.adjustTableColumns(tableWidth)

		// Update preview pane size
		previewWidth := availableWidth * 35 / 100
		m.previewPane.Width = previewWidth
		m.previewPane.Height = max(15, availableHeight-3)
	} else {
		// Full width table - use full space
		m.showPreview = false
		m.adjustTableColumns(availableWidth)
	}

	m.table.SetHeight(max(10, availableHeight))

	// Update viewport size for detail view
	m.viewport.Width = availableWidth
	m.viewport.Height = availableHeight

	// Update input widths
	inputWidth := max(30, availableWidth/2-10)
	m.searchInput.Width = inputWidth
	m.filterInput.Width = inputWidth
}

func (m *GitHubIssuesModel) adjustTableColumns(availableWidth int) {
	var columns []table.Column

	if availableWidth < 80 {
		// Very small terminal - minimal columns
		columns = []table.Column{
			{Title: "#", Width: 6},
			{Title: "Title", Width: availableWidth - 25},
			{Title: "State", Width: 8},
			{Title: "Assignee", Width: 10},
		}
	} else if availableWidth < 120 {
		// Medium terminal - reduce some columns
		columns = []table.Column{
			{Title: "#", Width: 7},
			{Title: "Title", Width: availableWidth - 55},
			{Title: "State", Width: 9},
			{Title: "Assignee", Width: 12},
			{Title: "Labels", Width: 15},
			{Title: "Updated", Width: 10},
		}
	} else {
		// Full size - all columns
		titleWidth := max(20, availableWidth-75) // Ensure minimum width
		columns = []table.Column{
			{Title: "#", Width: 7},
			{Title: "Title", Width: titleWidth},
			{Title: "State", Width: 9},
			{Title: "Assignee", Width: 15},
			{Title: "Labels", Width: 18},
			{Title: "Updated", Width: 10},
			{Title: "Repo", Width: 15},
		}
	}

	m.currentColumns = columns
	m.table.SetColumns(columns)

	// Update table width to match available width
	m.table.SetWidth(availableWidth)

	// Apply table styling every time columns change to ensure proper rendering
	tableStyles := table.DefaultStyles()
	tableStyles.Header = tableStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(primaryColor).
		BorderBottom(true).
		Bold(true).
		Foreground(primaryColor)

	tableStyles.Selected = tableStyles.Selected.
		Bold(true).
		Foreground(lipgloss.Color("#000000")).
		Background(primaryColor)

	m.table.SetStyles(tableStyles)

	// Force a complete refresh of the table data
	m.updateTableRows()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *GitHubIssuesModel) applyFilters() {
	m.filteredIssues = m.issues

	// Apply search filter
	searchTerm := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))
	if searchTerm != "" {
		var filtered []services.IssueWithRepo
		for _, issue := range m.filteredIssues {
			if m.matchesSearch(issue, searchTerm) {
				filtered = append(filtered, issue)
			}
		}
		m.filteredIssues = filtered
	}

	// Apply advanced filters
	filterTerm := strings.ToLower(strings.TrimSpace(m.filterInput.Value()))
	if filterTerm != "" {
		var filtered []services.IssueWithRepo
		for _, issue := range m.filteredIssues {
			if m.matchesFilter(issue, filterTerm) {
				filtered = append(filtered, issue)
			}
		}
		m.filteredIssues = filtered
	}

	// Update table rows
	m.updateTableRows()

	// Update preview if enabled and there are items
	if m.showPreview && len(m.filteredIssues) > 0 {
		cursor := m.table.Cursor()
		if cursor < len(m.filteredIssues) {
			m.updatePreviewPane(cursor)
		}
	}
}

func (m *GitHubIssuesModel) matchesSearch(issue services.IssueWithRepo, searchTerm string) bool {
	// Search in title
	if issue.Issue.Title != nil && strings.Contains(strings.ToLower(*issue.Issue.Title), searchTerm) {
		return true
	}

	// Search in body
	if issue.Issue.Body != nil && strings.Contains(strings.ToLower(*issue.Issue.Body), searchTerm) {
		return true
	}

	// Search in assignee
	if issue.Issue.Assignee != nil && issue.Issue.Assignee.Login != nil {
		if strings.Contains(strings.ToLower(*issue.Issue.Assignee.Login), searchTerm) {
			return true
		}
	}

	// Search in labels
	if issue.Issue.Labels != nil {
		for _, label := range issue.Issue.Labels {
			if label.Name != nil && strings.Contains(strings.ToLower(*label.Name), searchTerm) {
				return true
			}
		}
	}

	// Search in repository
	if strings.Contains(strings.ToLower(issue.Repo), searchTerm) {
		return true
	}

	return false
}

func (m *GitHubIssuesModel) matchesFilter(issue services.IssueWithRepo, filterTerm string) bool {
	// Parse filter format: "state:open label:bug author:username"
	filters := strings.Fields(filterTerm)

	for _, filter := range filters {
		parts := strings.SplitN(filter, ":", 2)
		if len(parts) != 2 {
			// Simple text search if no colon
			return m.matchesSearch(issue, filter)
		}

		key, value := parts[0], parts[1]

		switch key {
		case "state":
			if issue.Issue.State == nil || !strings.Contains(strings.ToLower(*issue.Issue.State), value) {
				return false
			}
		case "label":
			found := false
			if issue.Issue.Labels != nil {
				for _, label := range issue.Issue.Labels {
					if label.Name != nil && strings.Contains(strings.ToLower(*label.Name), value) {
						found = true
						break
					}
				}
			}
			if !found {
				return false
			}
		case "author", "assignee":
			if issue.Issue.Assignee == nil || issue.Issue.Assignee.Login == nil {
				return false
			}
			if !strings.Contains(strings.ToLower(*issue.Issue.Assignee.Login), value) {
				return false
			}
		case "repo":
			if !strings.Contains(strings.ToLower(issue.Repo), value) {
				return false
			}
		}
	}

	return true
}

func (m *GitHubIssuesModel) updateTableRows() {
	var rows []table.Row

	// Debug: Check if we have data
	if len(m.filteredIssues) == 0 {
		// Set empty state for table
		m.table.SetRows([]table.Row{})
		return
	}

	// Get current column configuration
	titleColWidth := 40 // Default fallback

	// Find the title column width
	for _, col := range m.currentColumns {
		if col.Title == "Title" {
			titleColWidth = col.Width
			break
		}
	}

	for _, issue := range m.filteredIssues {
		// Format issue number
		number := "N/A"
		if issue.Issue.Number != nil {
			number = fmt.Sprintf("#%d", *issue.Issue.Number)
		}

		// Format title with dynamic width
		title := "Untitled"
		if issue.Issue.Title != nil {
			title = *issue.Issue.Title
			// Smart truncation based on actual column width
			if len(title) > titleColWidth-3 {
				if len(title) > titleColWidth {
					// Show first part + "..." + last part for very long titles
					firstPart := (titleColWidth - 3) / 2
					lastPart := titleColWidth - 3 - firstPart
					if firstPart > 5 && lastPart > 5 {
						title = title[:firstPart] + "..." + title[len(title)-lastPart:]
					} else {
						title = title[:titleColWidth-3] + "..."
					}
				}
			}
		}

		// Format state with color
		state := "Unknown"
		if issue.Issue.State != nil {
			if *issue.Issue.State == "open" {
				state = "🟢 Open"
			} else {
				state = "🔴 Closed"
			}
		}

		// Format assignee
		assignee := "Unassigned"
		if issue.Issue.Assignee != nil && issue.Issue.Assignee.Login != nil {
			assignee = "@" + *issue.Issue.Assignee.Login
			// Truncate if needed based on available space
			maxAssigneeWidth := 13
			for _, col := range m.currentColumns {
				if col.Title == "Assignee" {
					maxAssigneeWidth = col.Width - 2
					break
				}
			}
			if len(assignee) > maxAssigneeWidth {
				assignee = assignee[:maxAssigneeWidth-3] + "..."
			}
		}

		// Format labels (only include if column exists)
		labels := ""
		hasLabelsCol := false
		maxLabelWidth := 15
		for _, col := range m.currentColumns {
			if col.Title == "Labels" {
				hasLabelsCol = true
				maxLabelWidth = col.Width - 2
				break
			}
		}

		if hasLabelsCol && issue.Issue.Labels != nil && len(issue.Issue.Labels) > 0 {
			var labelNames []string
			for i, label := range issue.Issue.Labels {
				if i >= 2 { // Show max 2 labels
					labelNames = append(labelNames, "...")
					break
				}
				if label.Name != nil {
					labelNames = append(labelNames, *label.Name)
				}
			}
			labels = strings.Join(labelNames, ",")
			if len(labels) > maxLabelWidth {
				labels = labels[:maxLabelWidth-3] + "..."
			}
		}

		// Format updated time (only include if column exists)
		updated := ""
		hasUpdatedCol := false
		for _, col := range m.currentColumns {
			if col.Title == "Updated" {
				hasUpdatedCol = true
				break
			}
		}

		if hasUpdatedCol {
			updated = "Unknown"
			if issue.Issue.UpdatedAt != nil {
				updated = issue.Issue.UpdatedAt.Format("Jan 02")
			}
		}

		// Format repository (only include if column exists)
		repo := ""
		hasRepoCol := false
		maxRepoWidth := 12
		for _, col := range m.currentColumns {
			if col.Title == "Repo" {
				hasRepoCol = true
				maxRepoWidth = col.Width - 2
				break
			}
		}

		if hasRepoCol {
			repo = issue.Repo
			if len(repo) > maxRepoWidth {
				repo = repo[:maxRepoWidth-3] + "..."
			}
		}

		// Build row based on available columns - MUST match column order exactly
		row := []string{number, title, state, assignee}
		if hasLabelsCol {
			row = append(row, labels)
		}
		if hasUpdatedCol {
			row = append(row, updated)
		}
		if hasRepoCol {
			row = append(row, repo)
		}

		rows = append(rows, row)
	}

	// Always set the rows, even if empty
	m.table.SetRows(rows)
}

func (m *GitHubIssuesModel) updateDetailView() {
	if m.selected == nil {
		return
	}

	issue := m.selected.Issue

	// Build detailed view content
	var content strings.Builder

	// Header with issue number and title
	header := fmt.Sprintf("#%d", *issue.Number)
	if issue.Title != nil {
		header += ": " + *issue.Title
	}
	content.WriteString(detailHeaderStyle.Render(header))
	content.WriteString("\n\n")

	// Metadata section
	metadata := []string{}

	if issue.State != nil {
		var stateStr string
		if *issue.State == "open" {
			stateStr = issueOpenStyle.Render("🟢 OPEN")
		} else {
			stateStr = issueClosedStyle.Render("🔴 CLOSED")
		}
		metadata = append(metadata, stateStr)
	}

	metadata = append(metadata, fmt.Sprintf("Repository: %s", m.selected.Repo))

	if issue.Assignee != nil && issue.Assignee.Login != nil {
		metadata = append(metadata, fmt.Sprintf("Assignee: @%s", *issue.Assignee.Login))
	} else {
		metadata = append(metadata, "Assignee: Unassigned")
	}

	if issue.CreatedAt != nil {
		metadata = append(metadata, fmt.Sprintf("Created: %s", issue.CreatedAt.Format("Jan 02, 2006")))
	}

	if issue.UpdatedAt != nil {
		metadata = append(metadata, fmt.Sprintf("Updated: %s", issue.UpdatedAt.Format("Jan 02, 2006")))
	}

	if issue.Comments != nil {
		metadata = append(metadata, fmt.Sprintf("Comments: %d", *issue.Comments))
	}

	content.WriteString(metaStyle.Render(strings.Join(metadata, " • ")))
	content.WriteString("\n\n")

	// Labels section
	if issue.Labels != nil && len(issue.Labels) > 0 {
		content.WriteString("Labels:\n")
		for _, label := range issue.Labels {
			if label.Name != nil {
				content.WriteString(labelStyle.Render(*label.Name))
			}
		}
		content.WriteString("\n\n")
	}

	// Description/Body section
	if issue.Body != nil && *issue.Body != "" {
		content.WriteString("Description:\n")
		content.WriteString(detailContentStyle.Render(*issue.Body))
	} else {
		content.WriteString(metaStyle.Render("No description provided."))
	}

	content.WriteString("\n\n")
	content.WriteString(metaStyle.Render("Press 'o' to open in browser • 'y' to copy description • 'c' to view comments • 'esc' to go back"))

	m.viewport.SetContent(content.String())
}

func (m *GitHubIssuesModel) updateDetailViewWithCommentInfo(commentInfo string) {
	// Update the detail view to include comment information
	m.updateDetailView()

	// Append the comment info to the viewport content
	currentContent := m.viewport.View()
	newContent := currentContent + "\n\n" + metaStyle.Render("💬 "+commentInfo)
	m.viewport.SetContent(newContent)
}

func (m *GitHubIssuesModel) View() string {
	if m.loading {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.spinner.View()+" Loading GitHub issues...",
			metaStyle.Render("This may take a moment..."),
		)
	}

	if m.loadingComments {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.spinner.View()+" Loading comments...",
			metaStyle.Render("Fetching comment thread..."),
		)
	}

	if m.error != "" {
		return lipgloss.NewStyle().
			Foreground(errorColor).
			Render("❌ Error: " + m.error)
	}

	switch m.currentView {
	case viewModeDetail:
		return m.renderDetailView()
	case viewModeComments:
		return m.renderCommentsView()
	default:
		return m.renderTableView()
	}
}

func (m *GitHubIssuesModel) renderTableView() string {
	var sections []string

	// Quick filters section
	if m.currentView == viewModeTable {
		var quickFilterButtons []string
		for i, filter := range quickFilters {
			style := quickFilterInactiveStyle
			if m.activeQuickFilter == i {
				style = quickFilterStyle
			}
			button := style.Render(fmt.Sprintf("%s %s", filter.key, filter.name))
			quickFilterButtons = append(quickFilterButtons, button)
		}

		quickFiltersRow := lipgloss.JoinHorizontal(lipgloss.Left, quickFilterButtons...)
		quickFiltersSection := filterBoxStyle.Render(
			"Quick Filters:\n" + quickFiltersRow + "\n" +
				metaStyle.Render("Press 1-6 to toggle • f: custom filter • s: search"),
		)
		sections = append(sections, quickFiltersSection)
	}

	// Custom filter section
	if m.showFilters {
		filterContent := lipgloss.JoinVertical(
			lipgloss.Left,
			"🔍 Search: "+m.searchInput.View(),
			"🎯 Filter: "+m.filterInput.View(),
			metaStyle.Render("Examples: state:open, label:bug, assignee:username, repo:Azure/AKS"),
		)
		filterSection := filterBoxStyle.Render(filterContent)
		sections = append(sections, filterSection)
	}

	// Table section with enhanced styling - ALWAYS render the table header and content
	tableHeader := headerStyle.Render("GitHub Issues")
	tableView := m.table.View()
	tableContent := tableHeader + "\n" + tableView

	// Check if we should show split view (table + preview)
	showSplitView := m.showPreview && len(m.filteredIssues) > 0 && m.width > 120

	if showSplitView {
		// Split view with table and preview
		cursor := m.table.Cursor()
		if cursor < len(m.filteredIssues) {
			m.updatePreviewPane(cursor)
		}

		previewHeader := headerStyle.Render("Preview")
		previewContent := previewHeader + "\n" + m.previewPane.View()

		// Calculate precise widths - ensure they add up correctly
		tableWidth := (m.width * 60) / 100
		previewWidth := (m.width * 35) / 100
		spacerWidth := m.width - tableWidth - previewWidth

		// Ensure minimum widths
		if tableWidth < 40 {
			tableWidth = 40
		}
		if previewWidth < 30 {
			previewWidth = 30
		}

		tableSection := lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(tableWidth).Render(tableContent),
			lipgloss.NewStyle().Width(spacerWidth).Render(""),
			lipgloss.NewStyle().Width(previewWidth).Render(previewContent),
		)
		sections = append(sections, tableSection)
	} else {
		// Full width table
		if m.showPreview && m.width <= 120 {
			// Show hint that preview is disabled due to terminal size
			sections = append(sections, metaStyle.Render("💡 Preview disabled - terminal too narrow (need >120 chars)"))
		}
		sections = append(sections, tableContent)
	}

	// Enhanced status section
	status := fmt.Sprintf("📊 %d of %d issues", len(m.filteredIssues), len(m.issues))

	if m.searchInput.Value() != "" || m.filterInput.Value() != "" || m.activeQuickFilter != -1 {
		status += " (filtered)"
	}

	if len(m.filteredIssues) > 0 {
		cursor := m.table.Cursor()
		if cursor < len(m.filteredIssues) {
			selected := m.filteredIssues[cursor]
			status += fmt.Sprintf(" • Row %d/%d • Issue #%d",
				cursor+1, len(m.filteredIssues), *selected.Issue.Number)

			// Add issue details in status
			if selected.Issue.Assignee != nil && selected.Issue.Assignee.Login != nil {
				status += fmt.Sprintf(" • @%s", *selected.Issue.Assignee.Login)
			}
			if selected.Issue.Comments != nil && *selected.Issue.Comments > 0 {
				status += fmt.Sprintf(" • %d comments", *selected.Issue.Comments)
			}
		}
	}

	statusSection := statusBarStyle.Render(status)
	sections = append(sections, statusSection)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *GitHubIssuesModel) renderDetailView() string {
	if m.selected == nil {
		return "No issue selected"
	}

	return m.viewport.View()
}

func (m *GitHubIssuesModel) renderCommentsView() string {
	if m.selected == nil {
		return "No issue selected"
	}

	var content strings.Builder

	// Header with issue info
	issue := m.selected.Issue
	header := fmt.Sprintf("#%d: %s", *issue.Number, *issue.Title)
	content.WriteString(detailHeaderStyle.Render(header))
	content.WriteString("\n")
	content.WriteString(metaStyle.Render(fmt.Sprintf("Repository: %s • %d comments", m.selected.Repo, len(m.comments))))
	content.WriteString("\n\n")

	if len(m.comments) == 0 {
		content.WriteString(metaStyle.Render("No comments on this issue."))
	} else {
		// Display each comment
		for i, comment := range m.comments {
			// Comment header with author and date
			author := "Unknown"
			if comment.User != nil && comment.User.Login != nil {
				author = *comment.User.Login
			}

			date := "Unknown date"
			if comment.CreatedAt != nil {
				date = comment.CreatedAt.Format("Jan 02, 2006 at 3:04 PM")
			}

			commentHeader := fmt.Sprintf("💬 %s commented on %s", author, date)
			content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(commentHeader))
			content.WriteString("\n")

			// Comment body
			if comment.Body != nil {
				body := *comment.Body
				// Wrap long lines
				wrapped := lipgloss.NewStyle().Width(80).Render(body)
				content.WriteString(detailContentStyle.Render(wrapped))
			}

			// Add separator between comments (except for the last one)
			if i < len(m.comments)-1 {
				content.WriteString("\n" + strings.Repeat("─", 80) + "\n\n")
			}
		}
	}

	content.WriteString("\n\n")
	content.WriteString(metaStyle.Render("Press 'esc' to go back • 'y' to copy description • 'o' to open in browser"))

	m.viewport.SetContent(content.String())

	return m.viewport.View()
}

func (m *GitHubIssuesModel) Refresh() tea.Cmd {
	m.loading = true
	return m.loadIssues()
}

func (m *GitHubIssuesModel) loadIssues() tea.Cmd {
	return func() tea.Msg {
		issues, err := m.services.GetGitHubIssues()
		if err != nil {
			return errorMsg{Error: err.Error()}
		}
		return issuesLoadedMsg{Issues: issues}
	}
}

// Messages
type issuesLoadedMsg struct {
	Issues []services.IssueWithRepo
}

type errorMsg struct {
	Error string
}

func (m *GitHubIssuesModel) openInBrowser() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil || m.selected.Issue.HTMLURL == nil {
			return browserActionMsg{
				success: false,
				message: "No URL available for this issue",
			}
		}

		url := *m.selected.Issue.HTMLURL
		var cmd *exec.Cmd

		// Cross-platform browser opening
		switch runtime.GOOS {
		case "darwin": // macOS
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		case "linux":
			cmd = exec.Command("xdg-open", url)
		default:
			return browserActionMsg{
				success: false,
				message: "Unsupported operating system",
			}
		}

		err := cmd.Start()
		if err != nil {
			return browserActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to open browser: %v", err),
			}
		}

		return browserActionMsg{
			success: true,
			message: "Opened in browser",
		}
	}
}

func (m *GitHubIssuesModel) loadComments() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return commentsActionMsg{
				success: false,
				message: "No issue selected",
			}
		}

		// Parse repository owner and name from the repo string
		repoParts := strings.Split(m.selected.Repo, "/")
		if len(repoParts) != 2 {
			return commentsActionMsg{
				success: false,
				message: "Invalid repository format",
			}
		}

		owner, repo := repoParts[0], repoParts[1]
		issueNumber := *m.selected.Issue.Number

		// Load comments from GitHub API
		comments, err := m.services.GetGitHubIssueComments(owner, repo, issueNumber)
		if err != nil {
			return commentsActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to load comments: %v", err),
			}
		}

		return commentsLoadedMsg{
			comments: comments,
		}
	}
}

// Message types for actions
type browserActionMsg struct {
	success bool
	message string
}

type commentsActionMsg struct {
	success bool
	message string
}

type commentsLoadedMsg struct {
	comments []*github.IssueComment
}

func (m *GitHubIssuesModel) updateCommentsView() {
	if m.selected == nil {
		return
	}

	var content strings.Builder

	// Header with issue info
	issue := m.selected.Issue
	header := fmt.Sprintf("#%d: %s", *issue.Number, *issue.Title)
	content.WriteString(detailHeaderStyle.Render(header))
	content.WriteString("\n")
	content.WriteString(metaStyle.Render(fmt.Sprintf("Repository: %s • %d comments", m.selected.Repo, len(m.comments))))
	content.WriteString("\n\n")

	if len(m.comments) == 0 {
		content.WriteString(metaStyle.Render("No comments on this issue."))
	} else {
		// Display each comment
		for i, comment := range m.comments {
			// Comment header with author and date
			author := "Unknown"
			if comment.User != nil && comment.User.Login != nil {
				author = *comment.User.Login
			}

			date := "Unknown date"
			if comment.CreatedAt != nil {
				date = comment.CreatedAt.Format("Jan 02, 2006 at 3:04 PM")
			}

			commentHeader := fmt.Sprintf("💬 %s commented on %s", author, date)
			content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(commentHeader))
			content.WriteString("\n")

			// Comment body
			if comment.Body != nil {
				body := *comment.Body
				// Wrap long lines
				wrapped := lipgloss.NewStyle().Width(80).Render(body)
				content.WriteString(detailContentStyle.Render(wrapped))
			}

			// Add separator between comments (except for the last one)
			if i < len(m.comments)-1 {
				content.WriteString("\n" + strings.Repeat("─", 80) + "\n\n")
			}
		}
	}

	content.WriteString("\n\n")
	content.WriteString(metaStyle.Render("Press 'esc' to go back • 'y' to copy description • 'o' to open in browser"))

	m.viewport.SetContent(content.String())
}

func (m *GitHubIssuesModel) updatePreviewPane(cursor int) {
	if cursor >= len(m.filteredIssues) {
		return
	}

	issue := m.filteredIssues[cursor]

	// Build preview pane content with full title prominence
	var content strings.Builder

	// Issue number - smaller, less prominent
	if issue.Issue.Number != nil {
		issueNum := metaStyle.Render(fmt.Sprintf("Issue #%d", *issue.Issue.Number))
		content.WriteString(issueNum + "\n")
	}

	// FULL TITLE - most prominent, wraps properly
	if issue.Issue.Title != nil {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Width(m.previewPane.Width - 4) // Use full preview width

		content.WriteString(titleStyle.Render(*issue.Issue.Title))
		content.WriteString("\n\n")
	}

	// Quick status line with key info
	var statusParts []string

	if issue.Issue.State != nil {
		if *issue.Issue.State == "open" {
			statusParts = append(statusParts, issueOpenStyle.Render("🟢 OPEN"))
		} else {
			statusParts = append(statusParts, issueClosedStyle.Render("🔴 CLOSED"))
		}
	}

	if issue.Issue.Assignee != nil && issue.Issue.Assignee.Login != nil {
		statusParts = append(statusParts, fmt.Sprintf("👤 @%s", *issue.Issue.Assignee.Login))
	} else {
		statusParts = append(statusParts, "👤 Unassigned")
	}

	if issue.Issue.Comments != nil && *issue.Issue.Comments > 0 {
		statusParts = append(statusParts, fmt.Sprintf("💬 %d", *issue.Issue.Comments))
	}

	if len(statusParts) > 0 {
		content.WriteString(strings.Join(statusParts, " • "))
		content.WriteString("\n\n")
	}

	// Repository and dates
	content.WriteString(metaStyle.Render(fmt.Sprintf("📁 %s", issue.Repo)))
	if issue.Issue.UpdatedAt != nil {
		content.WriteString(metaStyle.Render(fmt.Sprintf(" • 📅 %s", issue.Issue.UpdatedAt.Format("Jan 02, 2006"))))
	}
	content.WriteString("\n")

	// Labels - more compact and visual
	if issue.Issue.Labels != nil && len(issue.Issue.Labels) > 0 {
		content.WriteString("\n🏷️  ")
		for i, label := range issue.Issue.Labels {
			if i > 0 {
				content.WriteString(" ")
			}
			if label.Name != nil {
				content.WriteString(labelStyle.Render(*label.Name))
			}
		}
		content.WriteString("\n")
	}

	// Description preview - more readable
	if issue.Issue.Body != nil && *issue.Issue.Body != "" {
		body := *issue.Issue.Body

		// Better description preview - show more content
		descriptionStyle := lipgloss.NewStyle().
			Width(m.previewPane.Width - 4)

		if len(body) > 300 {
			body = body[:300] + "..."
		}

		content.WriteString("\n📝 Description:\n")
		content.WriteString(descriptionStyle.Render(body))
	}

	// Action hints at bottom
	content.WriteString("\n\n")
	hintStyle := lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Width(m.previewPane.Width - 4).
		Align(lipgloss.Center)

	content.WriteString(hintStyle.Render("↑↓ navigate • enter: full view • p: toggle • o: browser"))

	m.previewPane.SetContent(content.String())
}

func (m *GitHubIssuesModel) copyToClipboard() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil || m.selected.Issue.Body == nil {
			return browserActionMsg{
				success: false,
				message: "No description available to copy",
			}
		}

		body := *m.selected.Issue.Body
		err := copyToClipboard(body)
		if err != nil {
			return browserActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to copy to clipboard: %v", err),
			}
		}

		return browserActionMsg{
			success: true,
			message: "Description copied to clipboard",
		}
	}
}

func copyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}
