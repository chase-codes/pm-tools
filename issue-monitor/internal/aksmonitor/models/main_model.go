package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chase/pm-tools/issue-monitor/internal/aksmonitor/services"
)

type Tab int

const (
	TabGitHubIssues Tab = iota
	TabADOItems
	TabSyncOverview
	TabUpdatesFeed
	TabRoadmapReview
)

type MainModel struct {
	services      *services.Services
	currentTab    Tab
	githubIssues  *GitHubIssuesModel
	adoItems      *ADOItemsModel
	syncOverview  *SyncOverviewModel
	updatesFeed   *UpdatesFeedModel
	roadmapReview *RoadmapReviewModel
	loading       bool
	error         string
}

func NewMainModel(services *services.Services) *MainModel {
	return &MainModel{
		services:      services,
		currentTab:    TabGitHubIssues,
		githubIssues:  NewGitHubIssuesModel(services),
		adoItems:      NewADOItemsModel(services),
		syncOverview:  NewSyncOverviewModel(services),
		updatesFeed:   NewUpdatesFeedModel(services),
		roadmapReview: NewRoadmapReviewModel(services),
	}
}

func (m *MainModel) Init() tea.Cmd {
	return tea.Batch(
		m.githubIssues.Init(),
		m.adoItems.Init(),
		m.syncOverview.Init(),
		m.updatesFeed.Init(),
		m.roadmapReview.Init(),
	)
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Forward window size to all child models with adjusted dimensions
		// Reserve space for main header and footer
		adjustedHeight := msg.Height - 6 // Reserve space for title, tabs, separator, footer
		adjustedMsg := tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: adjustedHeight,
		}

		var cmds []tea.Cmd

		githubModel, cmd := m.githubIssues.Update(adjustedMsg)
		m.githubIssues = githubModel.(*GitHubIssuesModel)
		cmds = append(cmds, cmd)

		adoModel, cmd := m.adoItems.Update(adjustedMsg)
		m.adoItems = adoModel.(*ADOItemsModel)
		cmds = append(cmds, cmd)

		syncModel, cmd := m.syncOverview.Update(adjustedMsg)
		m.syncOverview = syncModel.(*SyncOverviewModel)
		cmds = append(cmds, cmd)

		updatesModel, cmd := m.updatesFeed.Update(adjustedMsg)
		m.updatesFeed = updatesModel.(*UpdatesFeedModel)
		cmds = append(cmds, cmd)

		roadmapModel, cmd := m.roadmapReview.Update(adjustedMsg)
		m.roadmapReview = roadmapModel.(*RoadmapReviewModel)
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "1":
			m.currentTab = TabGitHubIssues
			return m, nil
		case "2":
			m.currentTab = TabADOItems
			return m, nil
		case "3":
			m.currentTab = TabSyncOverview
			return m, nil
		case "4":
			m.currentTab = TabUpdatesFeed
			return m, nil
		case "5":
			m.currentTab = TabRoadmapReview
			return m, nil
		}
	case RefreshCmd:
		return m, m.refreshAll()
	case ErrorMsg:
		m.error = msg.Error
		return m, nil
	}

	// Delegate to current tab
	switch m.currentTab {
	case TabGitHubIssues:
		model, cmd := m.githubIssues.Update(msg)
		m.githubIssues = model.(*GitHubIssuesModel)
		return m, cmd
	case TabADOItems:
		model, cmd := m.adoItems.Update(msg)
		m.adoItems = model.(*ADOItemsModel)
		return m, cmd
	case TabSyncOverview:
		model, cmd := m.syncOverview.Update(msg)
		m.syncOverview = model.(*SyncOverviewModel)
		return m, cmd
	case TabUpdatesFeed:
		model, cmd := m.updatesFeed.Update(msg)
		m.updatesFeed = model.(*UpdatesFeedModel)
		return m, cmd
	case TabRoadmapReview:
		model, cmd := m.roadmapReview.Update(msg)
		m.roadmapReview = model.(*RoadmapReviewModel)
		return m, cmd
	}

	return m, nil
}

func (m *MainModel) View() string {
	if m.loading {
		return m.renderLoading()
	}

	header := m.renderHeader()
	content := m.renderContent()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

func (m *MainModel) renderHeader() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00ff00")).
		Render("AKS Networking PM Dashboard")

	// Show configured repositories
	repoInfo := ""
	if len(m.services.GetConfig().Repositories) > 0 {
		repoInfo = "Monitoring: "
		for i, repo := range m.services.GetConfig().Repositories {
			if i > 0 {
				repoInfo += ", "
			}
			repoInfo += repo.FullName()
		}
		repoInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Render(repoInfo)
	}

	tabs := []string{
		"1. GitHub Issues",
		"2. ADO Items",
		"3. Sync Overview",
		"4. Updates Feed",
		"5. Roadmap Review",
	}

	tabStyle := lipgloss.NewStyle().
		Padding(0, 1).
		MarginRight(1).
		Foreground(lipgloss.Color("#CCCCCC"))

	activeTabStyle := tabStyle.Copy().
		Background(lipgloss.Color("#00ff00")).
		Foreground(lipgloss.Color("#000000")).
		Bold(true)

	var tabViews []string
	for i, tab := range tabs {
		if Tab(i) == m.currentTab {
			tabViews = append(tabViews, activeTabStyle.Render(tab))
		} else {
			tabViews = append(tabViews, tabStyle.Render(tab))
		}
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Left, tabViews...)

	// Build header sections
	headerSections := []string{title}

	if repoInfo != "" {
		headerSections = append(headerSections, repoInfo)
	}

	headerSections = append(headerSections, tabBar)
	headerSections = append(headerSections, lipgloss.NewStyle().Render("─"))

	return lipgloss.JoinVertical(lipgloss.Left, headerSections...)
}

func (m *MainModel) renderContent() string {
	switch m.currentTab {
	case TabGitHubIssues:
		return m.githubIssues.View()
	case TabADOItems:
		return m.adoItems.View()
	case TabSyncOverview:
		return m.syncOverview.View()
	case TabUpdatesFeed:
		return m.updatesFeed.View()
	case TabRoadmapReview:
		return m.roadmapReview.View()
	default:
		return "Unknown tab"
	}
}

func (m *MainModel) renderFooter() string {
	help := "q: quit • r: refresh • 1-5: switch tabs"

	// Add GitHub-specific help when on GitHub tab
	if m.currentTab == TabGitHubIssues {
		help += " • 1-6: quick filters • ↑↓: navigate • p: preview • enter: details • f: filter • s: search"
	}

	if m.error != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0000")).
			Render("Error: " + m.error)
		return lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Render("─"),
			help,
			errorStyle,
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Render("─"),
		help,
	)
}

func (m *MainModel) renderLoading() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00ff00")).
		Render("Loading...")
}

func (m *MainModel) refreshAll() tea.Cmd {
	return tea.Batch(
		m.githubIssues.Refresh(),
		m.adoItems.Refresh(),
		m.syncOverview.Refresh(),
		m.updatesFeed.Refresh(),
	)
}

// Commands
type RefreshCmd struct{}
type ErrorMsg struct{ Error string }
