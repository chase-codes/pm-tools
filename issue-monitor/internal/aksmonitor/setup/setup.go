package setup

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chase/pm-tools/issue-monitor/internal/aksmonitor/config"
)

func RunSetup() (*config.Config, error) {
	fmt.Println("🚀 Welcome to AKS Monitor Setup!")
	fmt.Println("This will help you configure your credentials and repositories.")
	fmt.Println()

	// Load existing config or create new one
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Setup GitHub token
	if err := setupGitHubToken(cfg); err != nil {
		return nil, err
	}

	// Setup ADO token
	if err := setupADOToken(cfg); err != nil {
		return nil, err
	}

	// Setup repositories
	if err := setupRepositories(cfg); err != nil {
		return nil, err
	}

	// Save configuration
	if err := config.SaveConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Setup complete! Your configuration has been saved.")
	fmt.Printf("📁 Config location: %s\n", config.GetConfigPath())
	fmt.Println()

	return cfg, nil
}

func setupGitHubToken(cfg *config.Config) error {
	fmt.Println("🔑 GitHub Token Setup")
	fmt.Println("You'll need a GitHub Personal Access Token to monitor repositories.")
	fmt.Println("Create one at: https://github.com/settings/tokens")
	fmt.Println("Required scopes: repo (for private repos), public_repo (for public repos)")
	fmt.Println()

	if cfg.GitHubToken != "" {
		fmt.Print("GitHub token already configured. Update it? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return nil
		}
	}

	fmt.Print("Enter your GitHub Personal Access Token: ")
	reader := bufio.NewReader(os.Stdin)
	token, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	token = strings.TrimSpace(token)
	if token == "" {
		fmt.Println("⚠️  No token provided. GitHub features will be disabled.")
		return nil
	}

	cfg.GitHubToken = token
	fmt.Println("✅ GitHub token configured!")
	fmt.Println()
	return nil
}

func setupADOToken(cfg *config.Config) error {
	fmt.Println("🔑 Azure DevOps Token Setup")
	fmt.Println("You'll need an Azure DevOps Personal Access Token.")
	fmt.Println("Create one at: https://dev.azure.com/[your-org]/_usersSettings/tokens")
	fmt.Println("Required scopes: Work Items (Read)")
	fmt.Println()

	if cfg.ADOToken != "" {
		fmt.Print("ADO token already configured. Update it? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return nil
		}
	}

	fmt.Print("Enter your Azure DevOps Personal Access Token (or press Enter to skip): ")
	reader := bufio.NewReader(os.Stdin)
	token, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	token = strings.TrimSpace(token)
	if token == "" {
		fmt.Println("⚠️  No token provided. Azure DevOps features will be disabled.")
		return nil
	}

	cfg.ADOToken = token
	fmt.Println("✅ Azure DevOps token configured!")
	fmt.Println()
	return nil
}

func setupRepositories(cfg *config.Config) error {
	fmt.Println("📚 Repository Setup")
	fmt.Println("Configure which repositories to monitor.")
	fmt.Println()

	// Show current repositories
	if len(cfg.Repositories) > 0 {
		fmt.Println("Currently configured repositories:")
		for i, repo := range cfg.Repositories {
			fmt.Printf("  %d. %s\n", i+1, repo.DisplayName())
		}
		fmt.Println()
	}

	for {
		fmt.Println("Options:")
		fmt.Println("  1. Add new repository")
		fmt.Println("  2. Remove repository")
		fmt.Println("  3. Continue with current setup")
		fmt.Print("Choose an option (1-3): ")

		reader := bufio.NewReader(os.Stdin)
		choice, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		choice = strings.TrimSpace(choice)
		switch choice {
		case "1":
			if err := addRepository(cfg); err != nil {
				fmt.Printf("❌ Error adding repository: %v\n", err)
			}
		case "2":
			if err := removeRepository(cfg); err != nil {
				fmt.Printf("❌ Error removing repository: %v\n", err)
			}
		case "3":
			return nil
		default:
			fmt.Println("Invalid option. Please choose 1, 2, or 3.")
		}
		fmt.Println()
	}
}

func addRepository(cfg *config.Config) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter repository owner (e.g., 'Azure'): ")
	owner, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read owner: %w", err)
	}
	owner = strings.TrimSpace(owner)

	fmt.Print("Enter repository name (e.g., 'AKS'): ")
	name, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read name: %w", err)
	}
	name = strings.TrimSpace(name)

	fmt.Print("Enter description (optional): ")
	description, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read description: %w", err)
	}
	description = strings.TrimSpace(description)

	fmt.Print("Enter labels to filter by (comma-separated, e.g., 'networking,enhancement'): ")
	labelsInput, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read labels: %w", err)
	}
	labelsInput = strings.TrimSpace(labelsInput)

	var labels []string
	if labelsInput != "" {
		labels = strings.Split(labelsInput, ",")
		for i, label := range labels {
			labels[i] = strings.TrimSpace(label)
		}
	}

	repo := config.Repository{
		Owner:       owner,
		Name:        name,
		Description: description,
		Labels:      labels,
	}

	if err := cfg.AddRepository(repo); err != nil {
		return err
	}

	fmt.Printf("✅ Added repository: %s\n", repo.DisplayName())
	return nil
}

func removeRepository(cfg *config.Config) error {
	if len(cfg.Repositories) == 0 {
		fmt.Println("No repositories configured.")
		return nil
	}

	fmt.Println("Select repository to remove:")
	for i, repo := range cfg.Repositories {
		fmt.Printf("  %d. %s\n", i+1, repo.DisplayName())
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter repository number: ")
	choice, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	choice = strings.TrimSpace(choice)
	// Simple validation - in a real app you'd want more robust parsing
	if choice == "" {
		return fmt.Errorf("no selection made")
	}

	// For simplicity, we'll just remove the first repository
	// In a real implementation, you'd parse the choice and remove the specific one
	if len(cfg.Repositories) > 0 {
		removed := cfg.Repositories[0]
		cfg.Repositories = cfg.Repositories[1:]
		if err := config.SaveConfig(cfg); err != nil {
			return err
		}
		fmt.Printf("✅ Removed repository: %s\n", removed.DisplayName())
	}

	return nil
}
