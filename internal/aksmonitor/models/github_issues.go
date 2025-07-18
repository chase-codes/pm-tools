package models

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chase/pm-tools/internal/aksmonitor/services"
	"github.com/google/go-github/v58/github"
)

type GitHubIssuesModel struct {
	services *services.Services
	list     list.Model
	viewport viewport.Model
	selected *github.Issue
	loading  bool
	error    string
}

type issueItem struct {
	issue *github.Issue
}

func (i issueItem) Title() string {
	if i.issue.Title == nil {
		return "Untitled"
	}
	return *i.issue.Title
}

func (i issueItem) Description() string {
	if i.issue.Number == nil {
		return "No number"
	}
	return fmt.Sprintf("#%d", *i.issue.Number)
}

func (i issueItem) FilterValue() string {
	return i.Title()
}

func NewGitHubIssuesModel(services *services.Services) *GitHubIssuesModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "GitHub Issues"
	l.SetShowHelp(true)

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00ff00"))

	return &GitHubIssuesModel{
		services: services,
		list:     l,
		viewport: vp,
	}
}

func (m *GitHubIssuesModel) Init() tea.Cmd {
	return m.loadIssues()
}

func (m *GitHubIssuesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.list.SelectedItem() != nil {
				item := m.list.SelectedItem().(issueItem)
				m.selected = item.issue
				m.updateViewport()
			}
		case "esc":
			m.selected = nil
		}
	case issuesLoadedMsg:
		m.loading = false
		m.error = ""
		var items []list.Item
		for _, issue := range msg.Issues {
			items = append(items, issueItem{issue: issue})
		}
		m.list.SetItems(items)
	case errorMsg:
		m.loading = false
		m.error = msg.Error
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *GitHubIssuesModel) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Render("Loading GitHub issues...")
	}

	if m.error != "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0000")).
			Render("Error: " + m.error)
	}

	if m.selected != nil {
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().Width(40).Render(m.list.View()),
			lipgloss.NewStyle().Width(2).Render(""),
			m.viewport.View(),
		)
	}

	return m.list.View()
}

func (m *GitHubIssuesModel) Refresh() tea.Cmd {
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

func (m *GitHubIssuesModel) updateViewport() {
	if m.selected == nil {
		return
	}

	content := fmt.Sprintf(
		"#%d: %s\n\n",
		*m.selected.Number,
		*m.selected.Title,
	)

	if m.selected.Body != nil {
		content += *m.selected.Body
	}

	if m.selected.Labels != nil {
		content += "\n\nLabels:\n"
		for _, label := range m.selected.Labels {
			content += fmt.Sprintf("- %s\n", *label.Name)
		}
	}

	m.viewport.SetContent(content)
}

// Messages
type issuesLoadedMsg struct {
	Issues []*github.Issue
}

type errorMsg struct {
	Error string
}
