package cli

import (
	"fmt"
	"strings"
)

// Command represents a parsed command
type Command struct {
	Type      CommandType
	Args      []string
	Flags     map[string]bool
	IsURL     bool
	URLToClone string
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
)

// ParseArgs parses command line arguments
func ParseArgs(args []string) (*Command, error) {
	cmd := &Command{
		Type:  CommandNone,
		Args:  []string{},
		Flags: make(map[string]bool),
	}
	
	if len(args) == 0 {
		// No arguments - default to interactive mode
		return cmd, nil
	}
	
	// Check for flags
	for i, arg := range args {
		if arg == "-i" || arg == "--interactive" {
			cmd.Type = CommandInteractive
			return cmd, nil
		}
		if arg == "-h" || arg == "--help" {
			cmd.Type = CommandHelp
			return cmd, nil
		}
		if arg == "-v" || arg == "--version" {
			cmd.Type = CommandVersion
			return cmd, nil
		}
		if arg == "--force" {
			cmd.Flags["force"] = true
			// Remove from args
			args = append(args[:i], args[i+1:]...)
		}
	}
	
	// Check first argument
	firstArg := args[0]
	
	// Check if it's a URL (implicit clone)
	if isGitURL(firstArg) {
		cmd.Type = CommandClone
		cmd.IsURL = true
		cmd.URLToClone = firstArg
		return cmd, nil
	}
	
	// Parse commands
	switch firstArg {
	case "list":
		cmd.Type = CommandList
	case "update":
		cmd.Type = CommandUpdate
		if len(args) > 1 {
			cmd.Args = args[1:]
		}
	case "remove":
		cmd.Type = CommandRemove
		if len(args) > 1 {
			cmd.Args = args[1:]
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
  get-repo list                   List all repositories
  get-repo update                 Launch TUI in update mode
  get-repo update <repo>          Update specific repository
  get-repo remove                 Launch TUI in remove mode
  get-repo remove <repo> [--force] Remove specific repository

Options:
  -i, --interactive    Force interactive TUI mode
  -h, --help          Show this help message
  -v, --version       Show version information
  --force             Skip confirmation prompts

Examples:
  get-repo https://github.com/user/repo
  get-repo list
  get-repo update my-project
  get-repo remove old-project --force`
}