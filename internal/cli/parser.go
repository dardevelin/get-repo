package cli

import (
	"fmt"
	"strings"
)

// Command represents a parsed command
type Command struct {
	Type       CommandType
	Args       []string
	Flags      map[string]bool
	IsURL      bool
	URLToClone string
	CloneURLs  []string // For bulk clone
	CloneFile  string   // File path for bulk clone from file
}

// CommandType represents the type of command
type CommandType int

const (
	CommandNone CommandType = iota
	CommandList
	CommandUpdate
	CommandRemove
	CommandClone
	CommandHelp
	CommandVersion
	CommandInteractive
	CommandCompletion
)

// ParseArgs parses command line arguments
func ParseArgs(args []string) (*Command, error) {
	cmd := &Command{
		Type:      CommandNone,
		Args:      []string{},
		Flags:     make(map[string]bool),
		CloneURLs: []string{},
	}

	if len(args) == 0 {
		// No arguments - default to interactive mode
		return cmd, nil
	}

	// Process flags and collect remaining args
	var remainingArgs []string
	skipNext := false

	for i, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}

		switch arg {
		case "-i", "--interactive":
			cmd.Type = CommandInteractive
			return cmd, nil
		case "-h", "--help":
			cmd.Type = CommandHelp
			return cmd, nil
		case "-v", "--version":
			cmd.Type = CommandVersion
			return cmd, nil
		case "--force":
			cmd.Flags["force"] = true
		case "--cd":
			cmd.Flags["cd"] = true
		case "-f", "--file":
			if i+1 < len(args) {
				cmd.CloneFile = args[i+1]
				skipNext = true
			} else {
				return nil, fmt.Errorf("--file requires a file path")
			}
		default:
			remainingArgs = append(remainingArgs, arg)
		}
	}

	if len(remainingArgs) == 0 {
		// No arguments after flags - default to interactive mode
		return cmd, nil
	}

	// Check first remaining argument
	firstArg := remainingArgs[0]

	// Check if it's a URL (implicit clone)
	if isGitURL(firstArg) {
		cmd.Type = CommandClone
		cmd.IsURL = true
		// Collect all URLs for bulk clone
		for _, arg := range remainingArgs {
			if isGitURL(arg) {
				cmd.CloneURLs = append(cmd.CloneURLs, arg)
			}
		}
		// Keep single URL for backward compatibility
		if len(cmd.CloneURLs) > 0 {
			cmd.URLToClone = cmd.CloneURLs[0]
		}
		return cmd, nil
	}

	// Parse commands
	switch firstArg {
	case "list":
		cmd.Type = CommandList
	case "update":
		cmd.Type = CommandUpdate
		if len(remainingArgs) > 1 {
			cmd.Args = remainingArgs[1:]
		}
	case "remove":
		cmd.Type = CommandRemove
		if len(remainingArgs) > 1 {
			cmd.Args = remainingArgs[1:]
		}
	case "clone":
		cmd.Type = CommandClone
		// Collect all URLs after 'clone' command
		if len(remainingArgs) > 1 {
			for _, arg := range remainingArgs[1:] {
				if isGitURL(arg) {
					cmd.CloneURLs = append(cmd.CloneURLs, arg)
				}
			}
			if len(cmd.CloneURLs) > 0 {
				cmd.URLToClone = cmd.CloneURLs[0]
			}
		}
	case "completion":
		cmd.Type = CommandCompletion
		if len(remainingArgs) > 1 {
			cmd.Args = remainingArgs[1:]
		}
	default:
		return nil, fmt.Errorf("unknown command: %s", firstArg)
	}

	return cmd, nil
}

// isGitURL checks if a string looks like a git URL
func isGitURL(s string) bool {
	// Check for short notation (anything with : that's not a protocol)
	if strings.Contains(s, ":") && !strings.Contains(s, "://") && !strings.HasPrefix(s, "git@") {
		// Ensure it's not a Windows path (C:\...)
		if len(s) > 2 && s[1] == ':' && s[2] == '\\' {
			return false
		}
		// Check if it has the pattern prefix:path
		colonIndex := strings.Index(s, ":")
		if colonIndex > 0 && colonIndex < len(s)-1 {
			return true
		}
	}

	// HTTP(S) URLs
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return true
	}

	// SSH URLs
	if strings.HasPrefix(s, "git@") {
		return true
	}

	// SCP-style URLs (e.g., user@host:path)
	if strings.Contains(s, "@") && strings.Contains(s, ":") {
		return true
	}

	return false
}

// NeedsInteractiveTUI determines if the command should launch the TUI
func (c *Command) NeedsInteractiveTUI() bool {
	switch c.Type {
	case CommandNone, CommandInteractive:
		return true
	case CommandUpdate, CommandRemove:
		// Launch TUI if no specific repo specified
		return len(c.Args) == 0
	default:
		return false
	}
}

// GetHelpText returns the help text
func GetHelpText() string {
	return `get-repo - A beautiful TUI for managing git repositories

Usage:
  get-repo                        Launch interactive TUI
  get-repo <url>                  Clone a repository
  get-repo <url1> <url2> ...      Clone multiple repositories
  get-repo -f <file>              Clone repositories from file
  get-repo clone <url1> <url2>    Clone multiple repositories
  get-repo list                   List all repositories
  get-repo update                 Launch TUI in update mode
  get-repo update <repo>          Update specific repository
  get-repo remove                 Launch TUI in remove mode
  get-repo remove <repo> [--force] Remove specific repository
  get-repo completion <shell>     Generate shell completion scripts

URL Format:
  Full URLs:
    https://github.com/user/repo
    git@github.com:user/repo.git
  
  Short notation (fuzzy matching):
    gh:user/repo              → https://github.com/user/repo
    gl:user/repo              → https://gitlab.com/user/repo
    bb:user/repo              → https://bitbucket.org/user/repo
    
    github:user/repo          → https://github.com/user/repo
    gitlab:user/repo          → https://gitlab.com/user/repo
    bitbucket:user/repo       → https://bitbucket.org/user/repo
    
    git:user/repo             → https://github.com/user/repo
    gitl:user/repo            → https://gitlab.com/user/repo
    bit:user/repo             → https://bitbucket.org/user/repo

Options:
  -i, --interactive    Force interactive TUI mode
  -h, --help          Show this help message
  -v, --version       Show version information
  -f, --file <path>   Read repository URLs from file
  --force             Skip confirmation prompts
  --cd                Output repository path after clone/update (use with: cd $(get-repo <url> --cd))

Completion:
  get-repo completion bash        Generate bash completion
  get-repo completion zsh         Generate zsh completion  
  get-repo completion fish        Generate fish completion

Examples:
  get-repo gh:dardevelin/get-repo
  get-repo gh:user/repo1 gitlab:user/repo2
  cd $(get-repo gh:golang/go --cd)
  get-repo -f repos.txt
  get-repo list
  cd $(get-repo update my-project --cd)
  get-repo remove old-project --force
  
  # Short notation examples:
  get-repo gh:golang/go
  get-repo gitlab:gitlab-org/gitlab
  get-repo bitbucket:atlassian/localstack
  
  # File format for -f option (repos.txt):
  # Comments start with #
  # One URL per line (supports short notation)
  gh:user/repo1
  gitlab:user/repo2
  https://github.com/user/repo3
  
  # Install bash completion
  get-repo completion bash > ~/.bash_completion.d/get-repo
  
  # Install zsh completion
  get-repo completion zsh > ~/.oh-my-zsh/completions/_get-repo`
}
