# PM Tools

A collection of productivity tools for Product Managers and Engineering teams, built in Go.

## 🛠️ Tools

### 🐛 [Issue Monitor](./issue-monitor/)

A terminal-based dashboard for monitoring issues and work items across GitHub and Azure DevOps repositories.

**Key Features:**
- Multi-repository monitoring with label filtering
- Interactive setup wizard for credentials
- Real-time updates and caching
- Beautiful terminal user interface

**Quick Start:**
```bash
cd issue-monitor
go run cmd/aks-monitor/main.go
```

### 🔄 [Release Tracker](./release-tracker/) *(Coming Soon)*

Track releases across multiple projects and generate release notes.

### 📊 [Metrics Dashboard](./metrics-dashboard/) *(Coming Soon)*

Team metrics and KPIs visualization.

### 🗺️ [Roadmap Planner](./roadmap-planner/) *(Coming Soon)*

Planning and roadmap management tools.

## 🏗️ Project Structure

```
pm-tools/
├── issue-monitor/           # Issue monitoring dashboard
├── release-tracker/         # Release tracking tool (future)
├── metrics-dashboard/       # Metrics visualization (future)
├── roadmap-planner/         # Roadmap planning (future)
├── shared/                  # Shared utilities (future)
├── README.md               # This file
└── Makefile                # Main project Makefile
```

## 🚀 Development

Each tool is self-contained with its own:
- `go.mod` and `go.sum` files
- `Makefile` with build commands
- `README.md` with tool-specific documentation
- Independent dependencies

### Building All Tools

```bash
# Build all tools
make build-all

# Build specific tool
make build-issue-monitor

# See all available commands
make help
```

### Adding New Tools

1. Create a new directory for your tool
2. Initialize a new Go module: `go mod init github.com/chase/pm-tools/tool-name`
3. Add a `Makefile` and `README.md`
4. Update this main README

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- [Charm Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling library
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI components
- [Go GitHub](https://github.com/google/go-github) - GitHub API client
- [Azure DevOps Go API](https://github.com/microsoft/azure-devops-go-api) - ADO API client
