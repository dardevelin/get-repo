<div align="center">
  <img src="logo.svg" alt="get-repo logo" width="120" height="120">
  
  # get-repo
  
  **Manage all your git repositories from one place**
  
  [![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-blue.svg)](https://golang.org/)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
</div>

---

`get-repo` helps you organize and manage your local git repositories. Browse them in a tree view, update multiple repos at once, and clone new ones - all from a beautiful terminal interface.

## Features

- **Tree View**: See all your repos organized by provider (github.com, gitlab.com, etc.)
- **Bulk Operations**: Update or remove multiple repositories at once
- **Bulk Clone**: Clone multiple repos from command line or a file
- **Short Notation**: Fuzzy matching for providers - `gh:user/repo`, `gl:user/repo`, `bit:user/repo`
- **Smart Interface**: Works as both an interactive TUI and traditional CLI tool
- **Fast**: Parallel operations for cloning and updating
- **Shell Completion**: Smart tab completion for bash, zsh, and fish with fuzzy matching hints

## Installation

```bash
brew tap dardevelin/get-repo
brew install get-repo
```

Or build from source:
```bash
git clone https://github.com/dardevelin/get-repo.git
cd get-repo
make build
```

### Development

For development, you'll also need:
- `go-md2man` for man page generation: `brew install go-md2man`
- `golangci-lint` for linting (optional): `brew install golangci-lint`

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed setup instructions.

## Quick Start

```bash
# Launch the interactive UI
get-repo

# Clone a repository (with short notation)
get-repo gh:user/repo
get-repo https://github.com/user/repo

# Clone and change to directory
cd $(get-repo gh:golang/go --cd)

# Clone multiple repositories
get-repo gh:user/repo1 gitlab:org/repo2 https://github.com/user/repo3

# Clone from a file
get-repo -f repos.txt

# List all your repositories
get-repo list

# Update repositories
get-repo update                      # Interactive selection
get-repo update github.com/user/repo  # Specific repo
cd $(get-repo update github.com/user/repo --cd)  # Update and cd
```

### Bulk Clone from File

Create a file with repository URLs (supports short notation):
```
# repos.txt
gh:charmbracelet/bubbletea
gitlab:org/project
https://github.com/charmbracelet/bubbles

# Comments and empty lines are ignored
git@github.com:company/backend.git
bitbucket:team/frontend
```

Then clone them all:
```bash
get-repo -f repos.txt
```

## Keyboard Shortcuts

**Navigation**
- `↑`/`↓` - Move up/down
- `←`/`→` - Collapse/expand folders
- `/` - Filter repositories

**Actions**
- `Space` - Select/deselect
- `a` - Select all
- `n` - Deselect all  
- `c` - Clone new repository
- `u` - Update selected
- `r` - Remove selected
- `q` - Quit

## Configuration

On first run, get-repo will help you set up your repository directory. By default, it organizes repos like this:

```
~/dev/vcs-codebases/
├── github.com/
│   └── user/
│       └── repo/
├── gitlab.com/
│   └── org/
│       └── project/
└── bitbucket.com/
    └── team/
        └── repo/
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
