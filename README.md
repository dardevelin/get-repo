<div align="center">
  <img src="logo.svg" alt="get-repo logo" width="120" height="120">
  
  # get-repo
  
  **A beautiful, hierarchical TUI for managing your git repositories**
  
  [![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-blue.svg)](https://golang.org/)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
</div>

---

`get-repo` provides an elegant terminal interface to browse, clone, update, and manage all your local git repositories. It organizes repositories by VCS provider (github.com, gitlab.com, etc.) in an expandable tree structure while maintaining full command-line functionality.

## Recent Improvements

- **Refined Status Indicators**: Clean unicode symbols (✓ ✗ •••) replace emoji-style icons
- **Fixed UI Stability**: Title stays visible, no more UI shifting when selecting items
- **Improved Error Display**: Consolidated error reporting at bottom of screen
- **Better Selection Visuals**: Removed checkboxes in favor of color highlighting and arrows
- **Batch Operation Fixes**: All selected items are now properly processed
- **Inline Async Operations**: Operations run without switching screens

## Features

### 🌳 **Hierarchical Tree View**
- **VCS Organization:** Repositories grouped by provider (github.com, gitlab.com, bitbucket.com)
- **Expandable Nodes:** Use arrow keys to expand/collapse organizations and users
- **Expansion State Preservation:** Tree structure remains intact during operations
- **Visual Status Indicators:** Real-time operation feedback with refined status icons (✓ ✗ •••)

### 🚀 **Powerful Operations**
- **Batch Operations:** Select multiple repositories with space bar for bulk updates/removals
- **Individual Actions:** Quick single-repository operations
- **Smart Selection:** Hierarchical selection (select entire VCS provider or user/org)
- **Safe Removal:** Confirmation prompts prevent accidental deletions

### 🎨 **Beautiful Interface**
- **Modern Design:** Clean, minimalist interface with smart color usage
- **Visual Selection:** Color-coded highlighting with arrow indicators (▸)
- **Eza-style Colors:** Professional color scheme compatible with common terminal themes
- **Nerd Font Icons:** Rich iconography for better visual organization
- **Responsive Design:** Adapts to terminal size changes with reserved UI space
- **File Browser:** Integrated directory selection with visual validation

### 🛠 **Comprehensive CLI**
- **Interactive & Non-interactive Modes:** Works both as TUI and traditional CLI
- **Shell Completion:** Full autocompletion support for bash, zsh, and fish
- **URL Detection:** Smart git URL recognition for direct cloning
- **Setup Wizard:** Guided first-run configuration with file browser
- **Asynchronous Operations:** Non-blocking UI with inline status updates
- **Error Consolidation:** Clean error reporting in dedicated UI section

## Installation

### Option 1: Download Binary (Recommended)
Download the latest release from [GitHub Releases](https://github.com/dardevelin/get-repo/releases) and place it in your PATH.

### Option 2: Build from Source
```bash
git clone https://github.com/dardevelin/get-repo.git
cd get-repo
go build -o get-repo ./cmd/get-repo
sudo mv get-repo /usr/local/bin/
```

### Option 3: Homebrew (Coming Soon)
```bash
brew tap dardevelin/get-repo
brew install get-repo
```

## Usage

### Interactive Mode (TUI)
Launch the beautiful tree-based interface:
```bash
get-repo                    # Launch interactive TUI
get-repo --interactive      # Force interactive mode
```

### Command Line Interface
Use get-repo from the command line for scripting and automation:
```bash
# Repository management
get-repo list                           # List all repositories
get-repo update                         # Interactive update mode
get-repo update github.com/user/repo    # Update specific repository
get-repo remove                         # Interactive removal mode  
get-repo remove github.com/user/repo    # Remove specific repository

# Cloning
get-repo https://github.com/user/repo   # Clone repository
get-repo git@github.com:user/repo.git   # Clone via SSH

# Help and information
get-repo --help                         # Show help
get-repo --version                      # Show version
```

### Shell Completion
Enable autocompletion for your shell:

#### Bash
```bash
# Install completion
get-repo completion bash > ~/.bash_completion.d/get-repo

# Or for a single session
source <(get-repo completion bash)
```

#### Zsh
```bash
# Install completion
get-repo completion zsh > ~/.oh-my-zsh/completions/_get-repo

# Or add to your .zshrc
echo 'source <(get-repo completion zsh)' >> ~/.zshrc
```

#### Fish
```bash
# Install completion
get-repo completion fish > ~/.config/fish/completions/get-repo.fish
```

## Interactive Mode Keybindings

### Navigation
| Key | Action |
|-----|--------|
| `↑`/`↓` | Navigate up/down |
| `←`/`h` | Collapse current node |
| `→`/`l` | Expand current node |
| `/` | Enable filtering |

### Repository Operations
| Key | Action |
|-----|--------|
| `Space` | Toggle selection (for batch operations) |
| `a` | Select all repositories |
| `n` | Deselect all repositories |
| `c` | Clone new repository |
| `u` | Update selected repository(s) |
| `r` | Remove selected repository(s) |

### General
| Key | Action |
|-----|--------|
| `q`/`Esc` | Quit application |
| `Ctrl+C` | Force quit |

## Configuration

On first run, get-repo will launch a setup wizard to configure:
- **Configuration location**: Where to store settings
- **Repository directory**: Where your git repositories are organized  
- **Shell integration**: Environment variable setup for custom paths

Default structure:
```
~/dev/vcs-codebases/
├── github.com/
│   ├── user1/
│   │   ├── repo1/
│   │   └── repo2/
│   └── user2/
│       └── repo3/
├── gitlab.com/
│   └── organization/
│       └── project/
└── bitbucket.com/
    └── team/
        └── repository/
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
