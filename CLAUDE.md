# CLAUDE.md

## Project Overview

**svnty** is a terminal user interface (TUI) application for managing Subversion (SVN) repositories, built in Go using the Bubble Tea framework. It provides a Git-like staging workflow for SVN operations with an intuitive keyboard-driven interface.

### Key Features
- Interactive status view with expandable sections (Unversioned, Unstaged, Staged, Ignored, Issues)
- File staging/unstaging with SVN changelists
- Inline diff viewing with syntax highlighting
- Commit interface with message editing
- Real-time SVN command execution

## Architecture Overview

The application follows a clean architecture with clear separation of concerns:

### Core Components

- **`main.go`**: Entry point with CLI flag parsing and application bootstrapping
- **`app/`**: Main application model implementing the Bubble Tea pattern
  - Manages global state and coordinates between different modes (Status/Commit)
  - Handles window sizing and top-level message routing

- **`svn/`**: SVN service layer with interface-based design
  - `Service` interface defines all SVN operations
  - `RealService`: Production implementation executing actual SVN commands
  - `MockService`: Test implementation for development
  - Handles XML parsing of SVN command outputs
  - Manages diff caching and changelist operations

- **`status/`**: Status view model and navigation logic
  - Complex cursor navigation system for hierarchical content
  - Expandable sections and inline diffs
  - Element-based rendering (HeaderElem, PathElem, DiffElem, BlankElem)
  - Scroll management with proper viewport handling

- **`commit/`**: Commit interface model
  - Textarea-based commit message editing
  - Integration with SVN changelist commits

- **`info/`**: Repository information display
- **`styles/`**: Centralized UI styling and theming
- **`tui/`**: Shared TUI utilities and message types
- **`logging/`**: Structured logging setup

### Key Design Patterns

1. **Bubble Tea MVC**: Each component implements `Init()`, `Update(tea.Msg)`, and `View()` methods
2. **Command Pattern**: Operations return `tea.Cmd` functions for async execution
3. **Service Layer**: SVN operations abstracted behind interface for testability
4. **Message Passing**: Components communicate through typed messages (e.g., `FetchStatusMsg`, `CommitModeMsg`)

## Development Commands

### Build and Run
```bash
# Build the application
go build

# Run with default settings (uses hardcoded test path)
go run .

# Run with custom SVN repository path
go run . -path /path/to/svn/repo

# Run with mock SVN service for testing
go run . -mock

# Enable logging to file
go run . -log svnty.log
```

### Development Workflow
```bash
# Download dependencies
go mod tidy

# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Build for production
go build -ldflags="-s -w" .
```

### SVN Changelist Integration

The application uses SVN changelists to implement Git-like staging:
- **Staging**: `svn changelist staged <file>`
- **Unstaging**: `svn changelist --remove <file>`
- **Committing**: `svn commit --changelist staged`

### Navigation and Controls

In Status mode:
- `j/k`: Navigate up/down through sections and files
- `enter`: Expand/collapse sections
- `=`: Toggle inline diff view for files
- `s`: Stage current file/selection
- `u`: Unstage current file (only in Staged section)
- `c`: Switch to Commit mode
- `q`: Quit application

In Commit mode:
- `esc`: Return to Status mode
- `tab`: Submit commit
- Regular text editing in message area

## Important Implementation Details

### Cursor System
The status view uses a sophisticated cursor system (`status/status.go:37-49`) that tracks:
- `ElementType`: Header, Path, Diff line, or Blank
- `Section`: Which SVN status section (Unversioned, Unstaged, etc.)
- `PathIdx`: Index within section's file list
- `DiffLine`: Specific line number within expanded diff

### SVN Status Mapping
Files are categorized into sections based on SVN status (`svn/svn.go:108-130`):
- **Unversioned**: New files not tracked by SVN
- **Unstaged**: Modified/added/deleted files not in changelist
- **Staged**: Files in "staged" changelist ready for commit
- **Ignored**: Files matching SVN ignore patterns
- **Issues**: Conflicted, external, or obstructed files

### Mock Service
The mock service (`svn/mock.go`) provides realistic test data for UI development without requiring a real SVN repository.

## File Organization

```
svnty/
├── main.go              # Application entry point
├── go.mod              # Go module definition
├── app/app.go          # Main application model
├── svn/                # SVN service layer
│   ├── svn.go         # Real SVN implementation
│   └── mock.go        # Mock implementation
├── status/status.go    # Status view and navigation
├── commit/commit.go    # Commit interface
├── info/info.go        # Repository info display
├── styles/styles.go    # UI styling
├── tui/               # TUI utilities
└── logging/logger.go  # Logging setup
```

This architecture enables easy testing, clear separation of UI and business logic, and extensible SVN operations.
