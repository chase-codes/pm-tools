package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chase/pm-tools/issue-monitor/internal/aksmonitor/config"
	"github.com/chase/pm-tools/issue-monitor/internal/aksmonitor/models"
	"github.com/chase/pm-tools/issue-monitor/internal/aksmonitor/services"
)

type App struct {
	model    *models.MainModel
	program  *tea.Program
	services *services.Services
	config   *config.Config
}

func NewApp(cfg *config.Config) *App {
	// Initialize services
	svcs := services.NewServices(cfg)

	// Initialize main model
	model := models.NewMainModel(svcs)

	// Create program
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	return &App{
		model:    model,
		program:  program,
		services: svcs,
		config:   cfg,
	}
}

func (a *App) Run() error {
	// Start background polling
	go a.startPolling()

	// Run the program
	_, err := a.program.Run()
	return err
}

func (a *App) startPolling() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send refresh command to the model
			a.program.Send(models.RefreshCmd{})
		}
	}
}
