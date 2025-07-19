package models

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chase/pm-tools/issue-monitor/internal/aksmonitor/services"
)

type reviewMode int

const (
	reviewModeList reviewMode = iota
	reviewModeEdit
	reviewModeComplete
)

type RoadmapReviewModel struct {
	services      *services.Services
	list          list.Model
	viewport      viewport.Model
	descInput     textinput.Model
	statusInput   textinput.Model
	dateInput     textinput.Model
	currentMode   reviewMode
	items         []RoadmapItem
	currentIndex  int
	loading       bool
	error         string
	reviewSession ReviewSession
}

type RoadmapItem struct {
	ID          string
	ItemTitle   string
	ItemDesc    string
	Status      string
	LastUpdated time.Time
	Assignee    string
	Labels      []string
	URL         string
}

type ReviewSession struct {
	StartTime     time.Time
	ReviewedItems []string
	UpdatedItems  []string
	Notes         string
}

func (i RoadmapItem) Title() string { return i.ItemTitle }
func (i RoadmapItem) Description() string {
	status := i.Status
	if status == "" {
		status = "No Status"
	}
	lastUpdated := "Never"
	if !i.LastUpdated.IsZero() {
		lastUpdated = i.LastUpdated.Format("Jan 02, 2006")
	}
	return fmt.Sprintf("%s • Last updated: %s", status, lastUpdated)
}
func (i RoadmapItem) FilterValue() string { return i.ItemTitle }

func NewRoadmapReviewModel(services *services.Services) *RoadmapReviewModel {
	// Initialize list for roadmap items
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Roadmap Review - Monthly Update Session"
	l.SetShowHelp(true)

	// Initialize viewport for instructions/summary
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1)

	// Initialize input fields for editing
	descInput := textinput.New()
	descInput.Placeholder = "Enter updated description..."
	descInput.CharLimit = 500
	descInput.Width = 60

	statusInput := textinput.New()
	statusInput.Placeholder = "Status (In Progress, Blocked, Completed, etc.)"
	statusInput.CharLimit = 50
	statusInput.Width = 30

	dateInput := textinput.New()
	dateInput.Placeholder = "Target date (YYYY-MM-DD) or milestone..."
	dateInput.CharLimit = 30
	dateInput.Width = 30

	return &RoadmapReviewModel{
		services:    services,
		list:        l,
		viewport:    vp,
		descInput:   descInput,
		statusInput: statusInput,
		dateInput:   dateInput,
		currentMode: reviewModeList,
		reviewSession: ReviewSession{
			StartTime: time.Now(),
		},
	}
}

func (m *RoadmapReviewModel) Init() tea.Cmd {
	return m.loadRoadmapItems()
}

func (m *RoadmapReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.currentMode {
		case reviewModeList:
			return m.updateListMode(msg)
		case reviewModeEdit:
			return m.updateEditMode(msg)
		case reviewModeComplete:
			return m.updateCompleteMode(msg)
		}
	case roadmapItemsLoadedMsg:
		m.loading = false
		m.error = ""
		var items []list.Item
		for _, item := range msg.Items {
			items = append(items, item)
		}
		m.list.SetItems(items)
		m.items = msg.Items
	case roadmapErrorMsg:
		m.loading = false
		m.error = msg.Error
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *RoadmapReviewModel) updateListMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Start editing selected item
		if m.list.SelectedItem() != nil {
			item := m.list.SelectedItem().(RoadmapItem)
			m.currentIndex = m.list.Index()
			m.currentMode = reviewModeEdit

			// Pre-populate fields with current values
			m.descInput.SetValue(item.ItemDesc)
			m.statusInput.SetValue(item.Status)
			m.descInput.Focus()

			return m, nil
		}
	case "r":
		return m, m.loadRoadmapItems()
	case "s":
		// Complete review session
		m.currentMode = reviewModeComplete
		return m, nil
	}
	return m, nil
}

func (m *RoadmapReviewModel) updateEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel editing
		m.currentMode = reviewModeList
		m.descInput.Blur()
		m.statusInput.Blur()
		m.dateInput.Blur()
		return m, nil
	case "tab":
		// Cycle through input fields
		if m.descInput.Focused() {
			m.descInput.Blur()
			m.statusInput.Focus()
		} else if m.statusInput.Focused() {
			m.statusInput.Blur()
			m.dateInput.Focus()
		} else {
			m.dateInput.Blur()
			m.descInput.Focus()
		}
		return m, nil
	case "enter":
		// Save changes
		if m.currentIndex < len(m.items) {
			item := &m.items[m.currentIndex]
			item.ItemDesc = m.descInput.Value()
			item.Status = m.statusInput.Value()
			item.LastUpdated = time.Now()

			// Add to reviewed items
			m.reviewSession.ReviewedItems = append(m.reviewSession.ReviewedItems, item.ID)
			if m.descInput.Value() != "" || m.statusInput.Value() != "" {
				m.reviewSession.UpdatedItems = append(m.reviewSession.UpdatedItems, item.ID)
			}
		}

		m.currentMode = reviewModeList
		m.descInput.Blur()
		m.statusInput.Blur()
		m.dateInput.Blur()

		// TODO: Save changes to GitHub Projects API
		return m, nil
	}

	// Update the focused input
	var cmd tea.Cmd
	if m.descInput.Focused() {
		m.descInput, cmd = m.descInput.Update(msg)
	} else if m.statusInput.Focused() {
		m.statusInput, cmd = m.statusInput.Update(msg)
	} else if m.dateInput.Focused() {
		m.dateInput, cmd = m.dateInput.Update(msg)
	}

	return m, cmd
}

func (m *RoadmapReviewModel) updateCompleteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.currentMode = reviewModeList
		return m, nil
	}
	return m, nil
}

func (m *RoadmapReviewModel) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Foreground(primaryColor).
			Render("Loading roadmap items...")
	}

	if m.error != "" {
		return lipgloss.NewStyle().
			Foreground(errorColor).
			Render("Error: " + m.error)
	}

	switch m.currentMode {
	case reviewModeList:
		return m.renderListView()
	case reviewModeEdit:
		return m.renderEditView()
	case reviewModeComplete:
		return m.renderCompleteView()
	default:
		return "Unknown view mode"
	}
}

func (m *RoadmapReviewModel) renderListView() string {
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Render("🗺️ Monthly Roadmap Review")

	instructions := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render("Instructions: ↑↓ navigate • enter: edit item • r: refresh • s: complete session")

	reviewInfo := fmt.Sprintf("Session started: %s • Reviewed: %d items • Updated: %d items",
		m.reviewSession.StartTime.Format("Jan 02, 2006 3:04 PM"),
		len(m.reviewSession.ReviewedItems),
		len(m.reviewSession.UpdatedItems))

	reviewStatus := lipgloss.NewStyle().
		Foreground(accentColor).
		Render(reviewInfo)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		instructions,
		reviewStatus,
		"",
		m.list.View(),
	)
}

func (m *RoadmapReviewModel) renderEditView() string {
	if m.currentIndex >= len(m.items) {
		return "Invalid item selected"
	}

	item := m.items[m.currentIndex]

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Render(fmt.Sprintf("📝 Editing: %s", item.Title))

	instructions := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render("tab: next field • enter: save • esc: cancel")

	form := lipgloss.JoinVertical(
		lipgloss.Left,
		"Description:",
		m.descInput.View(),
		"",
		"Status:",
		m.statusInput.View(),
		"",
		"Target Date/Milestone:",
		m.dateInput.View(),
	)

	formBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Render(form)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		instructions,
		"",
		formBox,
	)
}

func (m *RoadmapReviewModel) renderCompleteView() string {
	duration := time.Since(m.reviewSession.StartTime)

	summary := fmt.Sprintf(`🎉 Review Session Complete!

Duration: %v
Items Reviewed: %d
Items Updated: %d

Summary:
• You've successfully reviewed %d roadmap items
• Made updates to %d items with new descriptions, statuses, or dates
• Session started at %s

Press enter or esc to return to the list.`,
		duration.Round(time.Minute),
		len(m.reviewSession.ReviewedItems),
		len(m.reviewSession.UpdatedItems),
		len(m.reviewSession.ReviewedItems),
		len(m.reviewSession.UpdatedItems),
		m.reviewSession.StartTime.Format("Jan 02, 2006 3:04 PM"))

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(successColor).
		Padding(2, 4).
		Render(summary)
}

func (m *RoadmapReviewModel) Refresh() tea.Cmd {
	return m.loadRoadmapItems()
}

func (m *RoadmapReviewModel) loadRoadmapItems() tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement GitHub Projects API integration
		// For now, return mock data
		items := []RoadmapItem{
			{
				ID:          "ROADMAP-001",
				ItemTitle:   "Azure CNI IPv6 Support",
				ItemDesc:    "Implement dual-stack IPv6/IPv4 support in Azure CNI",
				Status:      "In Progress",
				LastUpdated: time.Now().AddDate(0, 0, -15),
				Assignee:    "networking-team",
				Labels:      []string{"networking", "ipv6", "high-priority"},
				URL:         "https://github.com/Azure/azure-container-networking/issues/123",
			},
			{
				ID:          "ROADMAP-002",
				ItemTitle:   "Windows Server 2025 Support",
				ItemDesc:    "Add support for Windows Server 2025 nodes in AKS",
				Status:      "Planning",
				LastUpdated: time.Now().AddDate(0, 0, -30),
				Assignee:    "windows-team",
				Labels:      []string{"windows", "compatibility"},
				URL:         "https://github.com/Azure/AKS/issues/456",
			},
			{
				ID:          "ROADMAP-003",
				ItemTitle:   "Network Policy v2 GA",
				ItemDesc:    "General availability of Network Policy v2 with enhanced security features",
				Status:      "In Progress",
				LastUpdated: time.Now().AddDate(0, 0, -7),
				Assignee:    "security-team",
				Labels:      []string{"networking", "security", "ga"},
				URL:         "https://github.com/Azure/azure-container-networking/issues/789",
			},
		}

		return roadmapItemsLoadedMsg{Items: items}
	}
}

// Messages
type roadmapItemsLoadedMsg struct {
	Items []RoadmapItem
}

type roadmapErrorMsg struct {
	Error string
}
