% GET-REPO(1) get-repo 1.0.0
% Darcy Brás da Silva
% July 2025

# NAME

get-repo - manage git repositories from one place

# SYNOPSIS

**get-repo** [*OPTIONS*]

**get-repo** *URL* [*URL*...]

**get-repo** **-f** *FILE* [*URL*...]

**get-repo** *COMMAND* [*ARGS*]

# DESCRIPTION

**get-repo** helps you organize and manage your local git repositories. Browse them in a tree view, update multiple repos at once, and clone new ones - all from a beautiful terminal interface.

# OPTIONS

**-h**, **--help**
: Show help message and exit

**-v**, **--version**
: Show version information and exit

**-i**, **--interactive**
: Force interactive TUI mode

**-f**, **--file** *FILE*
: Read repository URLs from file (one per line)

**--force**
: Skip confirmation prompts

# COMMANDS

**list**
: List all repositories

**update** [*REPO*...]
: Update repositories. Without arguments, launches interactive mode

**remove** [*REPO*...] [**--force**]
: Remove repositories. Without arguments, launches interactive mode

**clone** *URL* [*URL*...]
: Clone one or more repositories

**completion** *SHELL*
: Generate shell completion script (bash, zsh, or fish)

# EXAMPLES

Launch interactive mode:
```
get-repo
```

Clone a single repository:
```
get-repo https://github.com/user/repo
```

Clone multiple repositories:
```
get-repo https://github.com/user/repo1 https://github.com/user/repo2
```

Clone from file:
```
get-repo -f repos.txt
```

Update specific repository:
```
get-repo update github.com/user/repo
```

# FILE FORMAT

When using **-f**, the file should contain one URL per line. Comments starting with # and empty lines are ignored:

```
# My repositories
https://github.com/user/repo1
https://github.com/user/repo2

# Work projects
git@github.com:company/backend.git
```

# INTERACTIVE MODE

**Navigation:**
- **↑/↓** - Move up/down
- **←/→** - Collapse/expand folders
- **/** - Filter repositories

**Actions:**
- **Space** - Select/deselect
- **a** - Select all
- **n** - Deselect all
- **c** - Clone new repository
- **u** - Update selected
- **r** - Remove selected
- **q** - Quit

# FILES

**~/.config/get-repo/config.json**
: Configuration file

**~/dev/vcs-codebases/**
: Default repository directory

# ENVIRONMENT

**GET_REPO_CONFIG**
: Override configuration file location

# EXIT STATUS

**0**
: Success

**1**
: General error

# SEE ALSO

**git**(1)

# BUGS

Report bugs at https://github.com/dardevelin/get-repo/issues