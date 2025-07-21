package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chase/pm-tools/issue-monitor/internal/aksmonitor/config"
	"github.com/google/go-github/v58/github"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
)

type IssueWithRepo struct {
	Issue *github.Issue
	Repo  string
}

type Services struct {
	githubClient *github.Client
	adoClient    *azuredevops.Connection
	config       *config.Config
}

func NewServices(cfg *config.Config) *Services {
	var githubClient *github.Client
	if cfg.GitHubToken != "" {
		ts := github.BasicAuthTransport{
			Username: "token",
			Password: cfg.GitHubToken,
		}
		githubClient = github.NewClient(ts.Client())
	}

	var adoClient *azuredevops.Connection
	if cfg.ADOToken != "" {
		adoClient = &azuredevops.Connection{
			AuthorizationString: fmt.Sprintf("Basic %s", cfg.ADOToken),
		}
	}

	// Ensure cache directory exists
	os.MkdirAll(cfg.CacheDir, 0755)

	return &Services{
		githubClient: githubClient,
		adoClient:    adoClient,
		config:       cfg,
	}
}

func (s *Services) GetGitHubIssues() ([]IssueWithRepo, error) {
	if s.githubClient == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}

	// Try to load from cache first
	cacheFile := filepath.Join(s.config.CacheDir, "github_issues.json")
	if data, err := os.ReadFile(cacheFile); err == nil {
		var issues []IssueWithRepo
		if json.Unmarshal(data, &issues) == nil {
			// Return cached data if it's recent enough (less than 5 minutes old)
			if stat, err := os.Stat(cacheFile); err == nil {
				if time.Since(stat.ModTime()) < 5*time.Minute {
					return issues, nil
				}
			}
		}
	}

	// Fetch from all configured repositories
	var allIssues []IssueWithRepo
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, repo := range s.config.Repositories {
		opts := &github.IssueListByRepoOptions{
			State: "open",
			ListOptions: github.ListOptions{
				PerPage: 50, // Reduced from 100 to avoid rate limits
			},
		}

		// Add labels filter if specified
		if len(repo.Labels) > 0 {
			opts.Labels = repo.Labels
		}

		issues, _, err := s.githubClient.Issues.ListByRepo(ctx, repo.Owner, repo.Name, opts)
		if err != nil {
			// Don't fail completely - just log and continue
			fmt.Printf("Warning: failed to fetch issues from %s/%s: %v\n", repo.Owner, repo.Name, err)
			continue
		}

		repoName := repo.FullName()
		for _, issue := range issues {
			allIssues = append(allIssues, IssueWithRepo{
				Issue: issue,
				Repo:  repoName,
			})
		}
	}

	// Cache the results
	if data, err := json.Marshal(allIssues); err == nil {
		// Ensure cache directory exists
		os.MkdirAll(filepath.Dir(cacheFile), 0755)
		os.WriteFile(cacheFile, data, 0644)
	}

	return allIssues, nil
}

func (s *Services) GetADOItems() ([]workitemtracking.WorkItem, error) {
	if s.adoClient == nil {
		return nil, fmt.Errorf("ADO client not initialized")
	}

	// Try to load from cache first
	cacheFile := filepath.Join(s.config.CacheDir, "ado_items.json")
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
	return os.RemoveAll(s.config.CacheDir)
}

func (s *Services) GetConfig() *config.Config {
	return s.config
}

func (s *Services) GetGitHubIssueComments(owner, repo string, issueNumber int) ([]*github.IssueComment, error) {
	if s.githubClient == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}

	ctx := context.Background()

	// Fetch comments for the specific issue
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100, // GitHub allows max 100 comments per page
		},
	}

	var allComments []*github.IssueComment

	// Handle pagination
	for {
		comments, resp, err := s.githubClient.Issues.ListComments(ctx, owner, repo, issueNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch comments: %v", err)
		}

		allComments = append(allComments, comments...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allComments, nil
}
