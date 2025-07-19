package main

import (
	"flag"
	"log"

	"github.com/chase/pm-tools/issue-monitor/internal/aksmonitor/app"
	"github.com/chase/pm-tools/issue-monitor/internal/aksmonitor/config"
	"github.com/chase/pm-tools/issue-monitor/internal/aksmonitor/setup"
	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command line flags
	setupFlag := flag.Bool("setup", false, "Run interactive setup to configure credentials and repositories")
	flag.Parse()

	// Setup logging
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	var cfg *config.Config
	var err error

	if *setupFlag {
		// Run interactive setup
		cfg, err = setup.RunSetup()
		if err != nil {
			log.Fatal("Setup failed:", err)
		}
	} else {
		// Load existing configuration
		cfg, err = config.LoadConfig()
		if err != nil {
			log.Fatal("Failed to load configuration:", err)
		}

		// Check if we need to run setup
		if cfg.GitHubToken == "" && cfg.ADOToken == "" {
			logrus.Info("No credentials found. Running setup...")
			cfg, err = setup.RunSetup()
			if err != nil {
				log.Fatal("Setup failed:", err)
			}
		}
	}

	// Create and run the application
	app := app.NewApp(cfg)

	if err := app.Run(); err != nil {
		log.Fatal("Error running application:", err)
	}
}
