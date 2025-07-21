package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chase/pm-tools/issue-monitor/internal/aksmonitor/services"
)

type SyncOverviewModel struct {
	services *services.Services
	loading  bool
	error    string
}

func NewSyncOverviewModel(services *services.Services) *SyncOverviewModel {
	return &SyncOverviewModel{
		services: services,
	}
}

func (m *SyncOverviewModel) Init() tea.Cmd {
	return m.loadSyncData()
}

func (m *SyncOverviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case syncDataLoadedMsg:
		m.loading = false
		m.error = ""
	case syncErrorMsg:
		m.loading = false
		m.error = msg.Error
	}
	return m, nil
}

func (m *SyncOverviewModel) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Render("Loading sync data...")
	}

	if m.error != "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0000")).
			Render("Error: " + m.error)
	}

	content := "Sync Overview\n\n"
	content += "GitHub Issues: 0 linked\n"
	content += "ADO Items: 0 linked\n"
	content += "Out of sync: 0 items\n\n"
	content += "Press 'r' to refresh sync status"

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00ff00")).
		Padding(1, 2).
		Render(content)
}

func (m *SyncOverviewModel) Refresh() tea.Cmd {
	return m.loadSyncData()
}

func (m *SyncOverviewModel) loadSyncData() tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement actual sync data loading
		return syncDataLoadedMsg{}
	}
}

// Messages
type syncDataLoadedMsg struct{}
type syncErrorMsg struct{ Error string }
