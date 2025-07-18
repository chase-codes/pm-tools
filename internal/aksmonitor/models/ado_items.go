package models

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chase/pm-tools/internal/aksmonitor/services"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
)

type ADOItemsModel struct {
	services *services.Services
	list     list.Model
	viewport viewport.Model
	selected *workitemtracking.WorkItem
	loading  bool
	error    string
}

type adoItem struct {
	item *workitemtracking.WorkItem
}

func (i adoItem) Title() string {
	if i.item.Fields == nil {
		return "Untitled"
	}
	if title, ok := (*i.item.Fields)["System.Title"].(string); ok {
		return title
	}
	return "Untitled"
}

func (i adoItem) Description() string {
	if i.item.Id == nil {
		return "No ID"
	}
	return fmt.Sprintf("ID: %d", *i.item.Id)
}

func (i adoItem) FilterValue() string {
	return i.Title()
}

func NewADOItemsModel(services *services.Services) *ADOItemsModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "ADO Work Items"
	l.SetShowHelp(true)

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00ff00"))

	return &ADOItemsModel{
		services: services,
		list:     l,
		viewport: vp,
	}
}

func (m *ADOItemsModel) Init() tea.Cmd {
	return m.loadItems()
}

func (m *ADOItemsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.list.SelectedItem() != nil {
				item := m.list.SelectedItem().(adoItem)
				m.selected = item.item
				m.updateViewport()
			}
		case "esc":
			m.selected = nil
		}
	case adoItemsLoadedMsg:
		m.loading = false
		m.error = ""
		var items []list.Item
		for _, item := range msg.Items {
			items = append(items, adoItem{item: &item})
		}
		m.list.SetItems(items)
	case adoErrorMsg:
		m.loading = false
		m.error = msg.Error
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *ADOItemsModel) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Render("Loading ADO items...")
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

func (m *ADOItemsModel) Refresh() tea.Cmd {
	return m.loadItems()
}

func (m *ADOItemsModel) loadItems() tea.Cmd {
	return func() tea.Msg {
		items, err := m.services.GetADOItems()
		if err != nil {
			return adoErrorMsg{Error: err.Error()}
		}
		return adoItemsLoadedMsg{Items: items}
	}
}

func (m *ADOItemsModel) updateViewport() {
	if m.selected == nil {
		return
	}

	title := "Untitled"
	state := "Unknown"

	if m.selected.Fields != nil {
		if t, ok := (*m.selected.Fields)["System.Title"].(string); ok {
			title = t
		}
		if s, ok := (*m.selected.Fields)["System.State"].(string); ok {
			state = s
		}
	}

	content := fmt.Sprintf(
		"ID: %d\nTitle: %s\nState: %s\n",
		*m.selected.Id,
		title,
		state,
	)

	m.viewport.SetContent(content)
}

// Messages
type adoItemsLoadedMsg struct {
	Items []workitemtracking.WorkItem
}

type adoErrorMsg struct {
	Error string
}
