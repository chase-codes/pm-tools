package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	GitHubToken  string       `json:"github_token"`
	ADOToken     string       `json:"ado_token"`
	Repositories []Repository `json:"repositories"`
	CacheDir     string       `json:"cache_dir"`
}

type Repository struct {
	Owner       string   `json:"owner"`
	Name        string   `json:"name"`
	Labels      []string `json:"labels,omitempty"`
	Description string   `json:"description,omitempty"`
}

func (r Repository) FullName() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

func (r Repository) DisplayName() string {
	if r.Description != "" {
		return fmt.Sprintf("%s (%s)", r.FullName(), r.Description)
	}
	return r.FullName()
}

func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	// Try to load existing config
	if data, err := os.ReadFile(configPath); err == nil {
		var config Config
		if err := json.Unmarshal(data, &config); err == nil {
			return &config, nil
		}
	}

	// Return default config
	return &Config{
		Repositories: []Repository{
			{
				Owner:       "Azure",
				Name:        "AKS",
				Labels:      []string{"networking"},
				Description: "Azure Kubernetes Service",
			},
		},
		CacheDir: filepath.Join(os.TempDir(), "aks-monitor-cache"),
	}, nil
}

func SaveConfig(config *Config) error {
	configPath := getConfigPath()

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return "aks-monitor-config.json"
	}
	return filepath.Join(homeDir, ".config", "aks-monitor", "config.json")
}

func getConfigPath() string {
	return GetConfigPath()
}

func (c *Config) AddRepository(repo Repository) error {
	// Check if repository already exists
	for _, existing := range c.Repositories {
		if existing.Owner == repo.Owner && existing.Name == repo.Name {
			return fmt.Errorf("repository %s/%s already exists", repo.Owner, repo.Name)
		}
	}

	c.Repositories = append(c.Repositories, repo)
	return SaveConfig(c)
}

func (c *Config) RemoveRepository(owner, name string) error {
	for i, repo := range c.Repositories {
		if repo.Owner == owner && repo.Name == name {
			c.Repositories = append(c.Repositories[:i], c.Repositories[i+1:]...)
			return SaveConfig(c)
		}
	}
	return fmt.Errorf("repository %s/%s not found", owner, name)
}

func (c *Config) GetRepository(owner, name string) *Repository {
	for _, repo := range c.Repositories {
		if repo.Owner == owner && repo.Name == name {
			return &repo
		}
	}
	return nil
}
