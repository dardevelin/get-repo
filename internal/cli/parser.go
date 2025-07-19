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

Options:
  -i, --interactive    Force interactive TUI mode
  -h, --help          Show this help message
  -v, --version       Show version information
  -f, --file <path>   Read repository URLs from file
  --force             Skip confirmation prompts

Completion:
  get-repo completion bash        Generate bash completion
  get-repo completion zsh         Generate zsh completion  
  get-repo completion fish        Generate fish completion

Examples:
  get-repo https://github.com/user/repo
  get-repo https://github.com/user/repo1 https://github.com/user/repo2
  get-repo -f repos.txt
  get-repo -f repos.txt https://github.com/user/extra-repo
  get-repo list
  get-repo update my-project
  get-repo remove old-project --force
  
  # File format for -f option (repos.txt):
  # Comments start with #
  # One URL per line
  https://github.com/user/repo1
  https://github.com/user/repo2
  
  # Install bash completion
  get-repo completion bash > ~/.bash_completion.d/get-repo
  
  # Install zsh completion
  get-repo completion zsh > ~/.oh-my-zsh/completions/_get-repo`
}
