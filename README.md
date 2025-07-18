# PM Tools - AKS Networking Dashboard

A terminal-based TUI dashboard built in Go using the Charm Bubble Tea framework for AKS Networking PM workflow. This dashboard serves as a "single pane of glass" for daily monitoring, triaging, and updating GitHub issues and Azure DevOps work items.

## Features

### 🔍 GitHub Issues Management
- Monitor open issues from Azure/AKS repository
- Filter by "networking" label and keywords (CNI, etc.)
- View full issue details with Markdown rendering
- Assign issues to team members
- Add comments and labels
- Update issue status (open/close)

### 📋 Azure DevOps Integration
- Track work items in AKS project
- Filter by status, assignee, and type
- Update work item fields and status
- Add comments and links to GitHub issues

### 🔄 Sync Overview
- Track linked GitHub-ADO items
- Highlight status mismatches
- Bidirectional sync capabilities
- Create missing items automatically

### 📰 Updates Feed
- Microsoft RSS feed integration for AKS updates
- Competitor monitoring (EKS, GKE)
- Diff view for changes

### ⚡ Performance Features
- Non-blocking API calls with goroutines
- Local JSON caching to reduce API calls
- Rate limit handling with retries
- Startup under 100ms
- Auto-refresh every 5 minutes

## Installation

### Prerequisites
- Go 1.21 or later
- GitHub Personal Access Token (PAT)
- Azure DevOps Personal Access Token (PAT)

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd pm-tools

# Install dependencies
make install

# Build the application
make build

# Run the application
make run
```

### Environment Variables

Set the following environment variables:

```bash
export GITHUB_TOKEN="your_github_pat_here"
export ADO_TOKEN="your_ado_pat_here"
```

## Usage

### Keyboard Navigation

| Key | Action |
|-----|--------|
| `1-4` | Switch between tabs |
| `j/k` | Navigate up/down in lists |
| `enter` | Select item or confirm action |
| `esc` | Go back or cancel |
| `r` | Refresh data |
| `q` | Quit application |

### Tab Overview

1. **GitHub Issues** - Monitor and manage GitHub issues
2. **ADO Items** - Track Azure DevOps work items  
3. **Sync Overview** - View synchronization status
4. **Updates Feed** - Latest updates and competitor info

### GitHub Issues Tab

- **List View**: Shows all open issues with networking label
- **Detail View**: Full issue body with Markdown rendering
- **Actions**: Assign, comment, label, close/reopen

### ADO Items Tab

- **List View**: Shows work items from AKS project
- **Detail View**: Work item details and description
- **Actions**: Update status, assign, add comments

### Sync Overview Tab

- **Status**: Shows linked items and sync status
- **Mismatches**: Highlights out-of-sync items
- **Actions**: Manual sync and create missing items

### Updates Feed Tab

- **AKS Updates**: Latest from Microsoft RSS feeds
- **Competitor Info**: EKS and GKE updates
- **Diff View**: Changes and new features

## Configuration

### Polling Interval

The default refresh interval is 5 minutes. To change this, modify the `startPolling()` function in `internal/aksmonitor/app/app.go`.

### Cache Settings

Cache files are stored in the system temp directory. To clear cache:

```bash
# The application will clear cache automatically, or you can:
rm -rf /tmp/aks-monitor-cache/
```

### API Rate Limits

The application handles GitHub and ADO API rate limits automatically with exponential backoff retries.

## Development

### Project Structure

```
pm-tools/
├── cmd/
│   └── aks-monitor/
│       └── main.go              # Application entry point
├── internal/
│   └── aksmonitor/
│       ├── app/
│       │   └── app.go           # Main application logic
│       ├── models/
│       │   ├── main_model.go    # Main UI model
│       │   ├── github_issues.go # GitHub issues model
│       │   ├── ado_items.go     # ADO items model
│       │   ├── sync_overview.go # Sync overview model
│       │   └── updates_feed.go  # Updates feed model
│       └── services/
│           └── services.go      # API services
├── go.mod                       # Go module file
├── Makefile                     # Build and development commands
└── README.md                    # This file
```

### Development Commands

```bash
# Install dependencies
make install

# Run in development mode (requires air)
make dev

# Format code
make fmt

# Run tests
make test

# Lint code
make lint

# Build for all platforms
make build-all
```

### Adding New Features

1. **New Tab**: Add to `Tab` enum in `main_model.go`
2. **New Model**: Create model file in `models/` directory
3. **New Service**: Add to `services.go`
4. **New API**: Implement in appropriate service

## Troubleshooting

### Common Issues

**GitHub API Errors**
- Verify `GITHUB_TOKEN` is set and valid
- Check token has appropriate permissions (repo, issues)
- Ensure token hasn't expired

**ADO API Errors**
- Verify `ADO_TOKEN` is set and valid
- Check token has work item read/write permissions
- Ensure project name is correct (hardcoded as "AKS")

**UI Not Responding**
- Check terminal supports UTF-8
- Ensure terminal is large enough (minimum 80x24)
- Try running with `TERM=xterm-256color`

### Debug Mode

To enable debug logging, modify the log level in `cmd/aks-monitor/main.go`:

```go
logrus.SetLevel(logrus.DebugLevel)
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Run `make test` and `make lint`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Charm Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling library
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI components
- [Go GitHub](https://github.com/google/go-github) - GitHub API client
- [Azure DevOps Go API](https://github.com/microsoft/azure-devops-go-api) - ADO API client
