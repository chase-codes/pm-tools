package main

import (
	"log"
	"os"

	"github.com/chase/pm-tools/internal/aksmonitor/app"
	"github.com/sirupsen/logrus"
)

func main() {
	// Setup logging
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Check for required environment variables
	githubToken := os.Getenv("GITHUB_TOKEN")
	adoToken := os.Getenv("ADO_TOKEN")

	if githubToken == "" {
		logrus.Warn("GITHUB_TOKEN environment variable not set. GitHub features will be disabled.")
	}

	if adoToken == "" {
		logrus.Warn("ADO_TOKEN environment variable not set. Azure DevOps features will be disabled.")
	}

	// Create and run the application
	app := app.NewApp(githubToken, adoToken)

	if err := app.Run(); err != nil {
		log.Fatal("Error running application:", err)
	}
}
