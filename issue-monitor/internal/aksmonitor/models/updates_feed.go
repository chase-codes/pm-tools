package models

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chase/pm-tools/issue-monitor/internal/aksmonitor/services"
)

type UpdatesFeedModel struct {
	services *services.Services
	viewport viewport.Model
	loading  bool
	error    string
}

func NewUpdatesFeedModel(services *services.Services) *UpdatesFeedModel {
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00ff00"))

	return &UpdatesFeedModel{
		services: services,
		viewport: vp,
	}
}

func (m *UpdatesFeedModel) Init() tea.Cmd {
	return m.loadUpdates()
}

func (m *UpdatesFeedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case updatesLoadedMsg:
		m.loading = false
		m.error = ""
		m.viewport.SetContent(msg.Content)
	case updatesErrorMsg:
		m.loading = false
		m.error = msg.Error
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *UpdatesFeedModel) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Render("Loading updates...")
	}

	if m.error != "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0000")).
			Render("Error: " + m.error)
	}

	return m.viewport.View()
}

func (m *UpdatesFeedModel) Refresh() tea.Cmd {
	return m.loadUpdates()
}

func (m *UpdatesFeedModel) loadUpdates() tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement actual RSS feed loading
		content := "Updates Feed\n\n"
		content += "Latest AKS Networking Updates:\n"
		content += "- No updates available\n\n"
		content += "Competitor Updates:\n"
		content += "- EKS: No recent updates\n"
		content += "- GKE: No recent updates\n\n"
		content += "Press 'r' to refresh"

		return updatesLoadedMsg{Content: content}
	}
}

// Messages
type updatesLoadedMsg struct {
	Content string
}

type updatesErrorMsg struct {
	Error string
}
