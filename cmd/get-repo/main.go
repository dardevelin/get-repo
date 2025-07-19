package main

import (
	"fmt"
	"os"

	"get-repo/config"
	"get-repo/internal/cli"
	"get-repo/internal/debug"
	"get-repo/internal/ui"
	"get-repo/pkg/version"

	tea "github.com/charmbracelet/bubbletea"
)

const bashCompletion = `#!/bin/bash

_get_repo_completion() {
    local cur prev opts repo_list
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Basic commands and options
    opts="list update remove clone completion --help --version --interactive --force --file"
    
    case "${prev}" in
        update|remove)
            # Get repository list for update/remove commands
            if command -v get-repo >/dev/null 2>&1; then
                repo_list=$(get-repo list 2>/dev/null)
                COMPREPLY=($(compgen -W "${repo_list}" -- ${cur}))
                return 0
            fi
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- ${cur}))
            return 0
            ;;
        --file|-f)
            # File completion
            COMPREPLY=($(compgen -f -- ${cur}))
            return 0
            ;;
        get-repo)
            # Complete with commands, URLs, or repository names
            COMPREPLY=($(compgen -W "${opts}" -- ${cur}))
            
            # Add URL completion hints
            if [[ ${cur} == http* ]] || [[ ${cur} == git@* ]]; then
                # Don't complete URLs, let user type them
                return 0
            fi
            
            return 0
            ;;
    esac
    
    # Default completion
    COMPREPLY=($(compgen -W "${opts}" -- ${cur}))
}

# Register the completion function
complete -F _get_repo_completion get-repo`

const zshCompletion = `#compdef get-repo

_get_repo() {
    local context state line
    typeset -A opt_args
    
    _arguments \
        '(-h --help)'{-h,--help}'[Show help message]' \
        '(-v --version)'{-v,--version}'[Show version information]' \
        '(-i --interactive)'{-i,--interactive}'[Force interactive TUI mode]' \
        '(-f --file)'{-f,--file}'[Read repository URLs from file]:file:_files' \
        '--force[Skip confirmation prompts]' \
        '*::command:_get_repo_command'
}

_get_repo_command() {
    local commands repos
    
    commands=(
        'list:List all repositories'
        'update:Update repositories'
        'remove:Remove repositories'
        'clone:Clone repositories'
        'completion:Generate shell completion scripts'
    )
    
    if (( CURRENT == 1 )); then
        # First argument: command or URL
        _describe -t commands 'command' commands
        _urls
    elif (( CURRENT >= 2 )); then
        case "$words[1]" in
            update|remove)
                # Get repository list
                if (( $+commands[get-repo] )); then
                    repos=(${(f)"$(get-repo list 2>/dev/null)"})
                    _describe -t repositories 'repository' repos
                fi
                ;;
            clone)
                # Multiple URLs can be provided
                _urls
                ;;
            completion)
                local shells=(bash zsh fish)
                _describe -t shells 'shell' shells
                ;;
            http*|git@*)
                # If first arg was a URL, continue accepting more URLs
                _urls
                ;;
        esac
    fi
}

_get_repo "$@"`

const fishCompletion = `# Fish completion for get-repo

# Basic commands
complete -c get-repo -f
complete -c get-repo -s h -l help -d "Show help message"
complete -c get-repo -s v -l version -d "Show version information"
complete -c get-repo -s i -l interactive -d "Force interactive TUI mode"
complete -c get-repo -s f -l file -r -d "Read repository URLs from file"
complete -c get-repo -l force -d "Skip confirmation prompts"

# Subcommands
complete -c get-repo -n "__fish_use_subcommand" -a "list" -d "List all repositories"
complete -c get-repo -n "__fish_use_subcommand" -a "update" -d "Update repositories"
complete -c get-repo -n "__fish_use_subcommand" -a "remove" -d "Remove repositories"
complete -c get-repo -n "__fish_use_subcommand" -a "clone" -d "Clone repositories"
complete -c get-repo -n "__fish_use_subcommand" -a "completion" -d "Generate shell completion scripts"

# Repository completion for update and remove
complete -c get-repo -n "__fish_seen_subcommand_from update remove" -a "(get-repo list 2>/dev/null)" -d "Repository"

# Shell completion for completion command
complete -c get-repo -n "__fish_seen_subcommand_from completion" -a "bash zsh fish" -d "Shell"

# Force flag for remove command
complete -c get-repo -n "__fish_seen_subcommand_from remove" -l force -d "Skip confirmation prompts"`

func main() {
	defer debug.LogFunction("main")()
	debug.Log("Application starting with args: %v", os.Args[1:])

	// Parse command line arguments
	cmd, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		debug.LogError(err, "parsing command line arguments")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Try 'get-repo --help' for more information.")
		os.Exit(1)
	}

	debug.Log("Parsed command type: %v", cmd.Type)

	// Handle help and version
	switch cmd.Type {
	case cli.CommandHelp:
		fmt.Println(cli.GetHelpText())
		return
	case cli.CommandVersion:
		fmt.Println(version.String())
		return
	case cli.CommandCompletion:
		if err := handleCompletion(cmd.Args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Load configuration
	debug.Log("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		debug.LogError(err, "loading configuration")
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	debug.Log("Configuration loaded: CodebasesPath=%s", cfg.CodebasesPath)

	// Check if we need setup
	if cfg.CodebasesPath == "" && cmd.Type != cli.CommandNone && cmd.Type != cli.CommandInteractive {
		fmt.Fprintln(os.Stderr, "Error: VCS_CODEBASES path not set.")
		fmt.Fprintln(os.Stderr, "Please run 'get-repo' interactively to configure.")
		os.Exit(1)
	}

	// Handle commands that need interactive TUI
	if cmd.NeedsInteractiveTUI() {
		debug.Log("Command needs interactive TUI, launching...")
		runTUI(getInitialState(cmd))
		return
	}

	// Handle non-interactive commands
	runner := cli.NewRunner(cfg)

	switch cmd.Type {
	case cli.CommandList:
		if err := runner.List(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case cli.CommandClone:
		// Handle bulk clone
		var urls []string

		// Check if we have a file to read from
		if cmd.CloneFile != "" {
			fileURLs, err := runner.ParseCloneFile(cmd.CloneFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
				os.Exit(1)
			}
			urls = append(urls, fileURLs...)
		}

		// Add any additional URLs from command line
		urls = append(urls, cmd.CloneURLs...)

		// If no URLs collected, fall back to single URL for backward compatibility
		if len(urls) == 0 && cmd.URLToClone != "" {
			urls = append(urls, cmd.URLToClone)
		}

		if len(urls) == 0 {
			fmt.Fprintln(os.Stderr, "Error: No URLs specified")
			os.Exit(1)
		}

		// Clone single or multiple repositories
		if len(urls) == 1 {
			if err := runner.Clone(urls[0]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := runner.CloneMultiple(urls); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

	case cli.CommandUpdate:
		if err := runner.Update(cmd.Args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case cli.CommandRemove:
		force := cmd.Flags["force"]
		if err := runner.Remove(cmd.Args, force); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command type: %v\n", cmd.Type)
		os.Exit(1)
	}
}

func getInitialState(cmd *cli.Command) ui.State {
	switch cmd.Type {
	case cli.CommandUpdate:
		return ui.StateUpdateSelection
	case cli.CommandRemove:
		return ui.StateRemoveSelection
	default:
		return ui.StateList
	}
}

func runTUI(initialState ui.State) {
	defer debug.LogFunction("runTUI")()
	debug.Log("Starting TUI with initial state: %v", initialState)

	// Set up logging for debugging
	if os.Getenv("DEBUG") != "" {
		debug.Log("DEBUG environment variable set, enabling tea logging")
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			debug.LogError(err, "setting up tea debug log")
			fmt.Fprintf(os.Stderr, "Error setting up debug log: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	// Create and run the program
	debug.Log("Creating UI model...")
	model := ui.InitialModel(initialState)
	debug.Log("UI model created successfully")

	debug.Log("Creating tea program...")
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	debug.Log("Tea program created, starting...")

	if _, err := p.Run(); err != nil {
		debug.LogError(err, "running tea program")
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
	debug.Log("Tea program finished successfully")
}

// handleCompletion generates and outputs shell completion scripts
func handleCompletion(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("completion command requires shell argument (bash, zsh, or fish)")
	}

	shell := args[0]
	switch shell {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}

	return nil
}
