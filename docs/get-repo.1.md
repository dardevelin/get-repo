% GET-REPO(1) get-repo 1.0.3
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

**--cd**
: Output repository path after clone/update (for use with command substitution)

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

# URL FORMAT

**get-repo** supports both full URLs and short notation for popular git hosting services:

**Full URLs:**
- `https://github.com/user/repo`
- `git@github.com:user/repo.git`
- `https://gitlab.com/user/repo`

**Short notation (with fuzzy matching):**
- `gh:user/repo` → `https://github.com/user/repo`
- `gl:user/repo` → `https://gitlab.com/user/repo`
- `bb:user/repo` → `https://bitbucket.org/user/repo`
- `github:user/repo` → `https://github.com/user/repo`
- `gitlab:user/repo` → `https://gitlab.com/user/repo`
- `bitbucket:user/repo` → `https://bitbucket.org/user/repo`
- `git:user/repo` → `https://github.com/user/repo` (defaults to GitHub)
- `gitl:user/repo` → `https://gitlab.com/user/repo`
- `bit:user/repo` → `https://bitbucket.org/user/repo`

# EXAMPLES

Launch interactive mode:
```
get-repo
```

Clone using short notation:
```
get-repo gh:dardevelin/get-repo
```

Clone multiple repositories with mixed notation:
```
get-repo gh:user/repo1 gitlab:user/repo2 https://github.com/user/repo3
```

Clone and change to directory:
```
cd $(get-repo gh:golang/go --cd)
```

Clone from file:
```
get-repo -f repos.txt
```

Update and change to directory:
```
cd $(get-repo update github.com/user/repo --cd)
```

# FILE FORMAT

When using **-f**, the file should contain one URL per line. Both full URLs and short notation are supported. Comments starting with # and empty lines are ignored:

```
# My repositories
gh:user/repo1
gitlab:user/repo2
https://github.com/user/repo3

# Work projects
git@github.com:company/backend.git
bitbucket:team/frontend
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