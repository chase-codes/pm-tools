package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-github/v58/github"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
)

type Services struct {
	githubClient *github.Client
	adoClient    *azuredevops.Connection
	cacheDir     string
}

func NewServices(githubToken, adoToken string) *Services {
	var githubClient *github.Client
	if githubToken != "" {
		ts := github.BasicAuthTransport{
			Username: "token",
			Password: githubToken,
		}
		githubClient = github.NewClient(ts.Client())
	}

	var adoClient *azuredevops.Connection
	if adoToken != "" {
		adoClient = &azuredevops.Connection{
			AuthorizationString: fmt.Sprintf("Basic %s", adoToken),
		}
	}

	cacheDir := filepath.Join(os.TempDir(), "aks-monitor-cache")
	os.MkdirAll(cacheDir, 0755)

	return &Services{
		githubClient: githubClient,
		adoClient:    adoClient,
		cacheDir:     cacheDir,
	}
}

func (s *Services) GetGitHubIssues() ([]*github.Issue, error) {
	if s.githubClient == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}

	// Try to load from cache first
	cacheFile := filepath.Join(s.cacheDir, "github_issues.json")
	if data, err := os.ReadFile(cacheFile); err == nil {
		var issues []*github.Issue
		if json.Unmarshal(data, &issues) == nil {
			return issues, nil
		}
	}

	// Fetch from API
	ctx := context.Background()
	issues, _, err := s.githubClient.Issues.ListByRepo(ctx, "Azure", "AKS", &github.IssueListByRepoOptions{
		State:  "open",
		Labels: []string{"networking"},
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})

	if err != nil {
		return nil, err
	}

	// Cache the results
	if data, err := json.Marshal(issues); err == nil {
		os.WriteFile(cacheFile, data, 0644)
	}

	return issues, nil
}

func (s *Services) GetADOItems() ([]workitemtracking.WorkItem, error) {
	if s.adoClient == nil {
		return nil, fmt.Errorf("ADO client not initialized")
	}

	// Try to load from cache first
	cacheFile := filepath.Join(s.cacheDir, "ado_items.json")
	if data, err := os.ReadFile(cacheFile); err == nil {
		var items []workitemtracking.WorkItem
		if json.Unmarshal(data, &items) == nil {
			return items, nil
		}
	}

	// For now, return mock data since ADO API setup is complex
	// In a real implementation, you would use the ADO client to fetch work items
	mockItems := []workitemtracking.WorkItem{
		{
			Id: github.Int(1),
			Fields: &map[string]interface{}{
				"System.Title": "AKS CNI Performance Issue",
				"System.State": "New",
			},
		},
		{
			Id: github.Int(2),
			Fields: &map[string]interface{}{
				"System.Title": "Network Policy Implementation",
				"System.State": "In Progress",
			},
		},
	}

	// Cache the results
	if data, err := json.Marshal(mockItems); err == nil {
		os.WriteFile(cacheFile, data, 0644)
	}

	return mockItems, nil
}

func (s *Services) UpdateGitHubIssue(number int, update *github.IssueRequest) error {
	if s.githubClient == nil {
		return fmt.Errorf("GitHub client not initialized")
	}

	ctx := context.Background()
	_, _, err := s.githubClient.Issues.Edit(ctx, "Azure", "AKS", number, update)
	return err
}

func (s *Services) AddGitHubComment(number int, comment string) error {
	if s.githubClient == nil {
		return fmt.Errorf("GitHub client not initialized")
	}

	ctx := context.Background()
	_, _, err := s.githubClient.Issues.CreateComment(ctx, "Azure", "AKS", number, &github.IssueComment{
		Body: &comment,
	})
	return err
}

func (s *Services) ClearCache() error {
	return os.RemoveAll(s.cacheDir)
}
