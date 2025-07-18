package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chase/pm-tools/internal/aksmonitor/models"
	"github.com/chase/pm-tools/internal/aksmonitor/services"
)

type App struct {
	model    *models.MainModel
	program  *tea.Program
	services *services.Services
}

func NewApp(githubToken, adoToken string) *App {
	// Initialize services
	svcs := services.NewServices(githubToken, adoToken)

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
