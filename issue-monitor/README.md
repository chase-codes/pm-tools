# AKS Monitor

A terminal-based dashboard for monitoring Azure Kubernetes Service (AKS) related issues and work items across GitHub and Azure DevOps.

## ✨ Features

- 📊 **Multi-repository monitoring**: Monitor issues from multiple GitHub repositories
- 🔍 **Label filtering**: Filter issues by specific labels (e.g., "networking", "enhancement")
- 🔄 **Real-time updates**: Automatic refresh every 5 minutes
- 💾 **Caching**: Local cache for faster loading
- 🎨 **Beautiful TUI**: Terminal user interface built with Bubble Tea
- 🔧 **Interactive setup**: Guided configuration wizard

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or later
- GitHub Personal Access Token (for GitHub features)
- Azure DevOps Personal Access Token (for ADO features)

### First Run Setup

Run the application for the first time to start the interactive setup:

```bash
cd issue-monitor
go run cmd/aks-monitor/main.go
```

The setup wizard will guide you through:
1. **GitHub Token Configuration**: Create a Personal Access Token at https://github.com/settings/tokens
   - Required scopes: `repo` (for private repos), `public_repo` (for public repos)
2. **Azure DevOps Token Configuration**: Create a Personal Access Token at https://dev.azure.com/[your-org]/_usersSettings/tokens
   - Required scopes: Work Items (Read)
3. **Repository Configuration**: Add repositories to monitor with optional label filters

### Manual Setup

If you prefer to run setup manually:

```bash
cd issue-monitor
go run cmd/aks-monitor/main.go -setup
```

### Building and Running

```bash
# Build the application
make build

# Run the built application
./bin/aks-monitor

# Or run directly with Go
go run cmd/aks-monitor/main.go
```

## ⚙️ Configuration

The application stores configuration in `~/.config/aks-monitor/config.json`. You can:

- **Add repositories**: Use the setup wizard or edit the config file directly
- **Update credentials**: Re-run setup with `-setup` flag
- **Configure labels**: Specify labels to filter issues (e.g., "networking", "enhancement")

### Example Configuration

```json
{
  "github_token": "ghp_your_token_here",
  "ado_token": "your_ado_token_here",
  "repositories": [
    {
      "owner": "Azure",
      "name": "AKS",
      "labels": ["networking"],
      "description": "Azure Kubernetes Service"
    },
    {
      "owner": "kubernetes",
      "name": "kubernetes",
      "labels": ["area/networking", "kind/feature"],
      "description": "Kubernetes"
    }
  ],
  "cache_dir": "/tmp/aks-monitor-cache"
}
```

## 🎮 Usage

### Navigation

- **1-4**: Switch between tabs (GitHub Issues, ADO Items, Sync Overview, Updates Feed)
- **Enter**: View issue details
- **Esc**: Return to issue list
- **r**: Refresh data
- **q**: Quit

### Repository Monitoring

The application monitors all configured repositories and displays:
- Issue number and title
- Repository name in brackets (e.g., "#123 [Azure/AKS]")
- Labels and descriptions
- Full issue body when selected

### Tab Overview

1. **GitHub Issues**: View and manage GitHub issues from configured repositories
2. **ADO Items**: Track Azure DevOps work items
3. **Sync Overview**: Monitor synchronization status between GitHub and ADO
4. **Updates Feed**: Latest updates and competitor information

## 🛠️ Development

### Building

```bash
make build
```

### Running Tests

```bash
make test
```

### Development Mode (with hot reload)

```bash
make dev  # Requires air to be installed
```

### Code Quality

```bash
make fmt   # Format code
make lint  # Lint code
```

### Project Structure

```
issue-monitor/
├── cmd/aks-monitor/          # Main application entry point
├── internal/aksmonitor/
│   ├── app/                  # Application logic
│   ├── config/               # Configuration management
│   ├── models/               # UI models and components
│   ├── services/             # External API services
│   └── setup/                # Interactive setup wizard
├── go.mod                    # Go module file
├── go.sum                    # Dependency checksums
├── Makefile                  # Build and development commands
└── README.md                 # This file
```

## 🔧 Troubleshooting

### Common Issues

**GitHub API Errors**
- Verify your GitHub token is valid and has the correct scopes
- Check that the repositories you're monitoring are accessible

**ADO API Errors**
- Verify your ADO token is valid and has work item read permissions
- Ensure your organization and project names are correct

**Configuration Issues**
- Delete `~/.config/aks-monitor/config.json` and re-run setup
- Check file permissions on the config directory

### Debug Mode

To enable debug logging, modify the log level in `cmd/aks-monitor/main.go`:

```go
logrus.SetLevel(logrus.DebugLevel)
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details. 