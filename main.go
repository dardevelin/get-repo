package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"get-repo/config"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateList state = iota
	stateSetup
	stateClone
	stateCloning
	stateUpdate
	stateRemoveConfirm
	stateUpdateSelection // New state for multi-selection update
	stateRemoveSelection // New state for multi-selection remove
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#5D5D5D")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	
)

type model struct {
	state     state
	config    config.Config
	list      list.Model
	textInput textinput.Model
	spinner   spinner.Model
	statusMsg string
	err       error
	selected  map[int]struct{} // For multi-selection
}

func initialModel(initialAppState state) model {
	cfg, err := config.Load()
	if err != nil {
		return model{err: err}
	}

	if cfg.CodebasesPath == "" {
		// If the path is not set, we must start in setup mode.
		ti := textinput.New()
		ti.Placeholder = "$HOME/dev/vcs-codebases"
		ti.Focus()
		ti.CharLimit = 156
		ti.Width = 50

		return model{
			state:     stateSetup,
			config:    cfg,
			textInput: ti,
		}
	}

	// Config is ready, scan for repos
	repos, err := scanForRepos(cfg.CodebasesPath)
	if err != nil {
		return model{err: err}
	}

	items := make([]list.Item, len(repos))
	for i, repo := range repos {
		items[i] = item(repo)
	}

	l := list.New(items, list.NewDefaultDelegate(), 1, 1)
	l.Title = "Your Repositories"
	l.SetShowHelp(false) // We'll render our own help

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Set list options based on initialAppState
	switch initialAppState {
	case stateUpdateSelection:
		l.Title = "Select repositories to update (Space to toggle, Enter to confirm)"
	case stateRemoveSelection:
		l.Title = "Select repositories to remove (Space to toggle, Enter to confirm)"
	}

	return model{
		state:   initialAppState,
		config:  cfg,
		list:    l,
		spinner: s,
		selected: make(map[int]struct{}),
	}
}

type item string

func (i item) Title() string       { return string(i) }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return string(i) }

// --- Messages ---
type cloneFinishedMsg struct{ err error }
type updateFinishedMsg struct{ err error }
type removeFinishedMsg struct{ err error }

// --- Commands ---
func (m model) cloneRepo(url string) tea.Cmd {
	return func() tea.Msg {
		clonePath := getClonePath(url)
		dest := filepath.Join(m.config.CodebasesPath, clonePath)

		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return cloneFinishedMsg{err: fmt.Errorf("failed to create directory: %w", err)}
		}

		cmd := exec.Command("git", "clone", url, dest)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return cloneFinishedMsg{err: fmt.Errorf("git clone failed: %w", err)}
		}
		return cloneFinishedMsg{err: nil}
	}
}

func (m model) updateRepo(repoName string) tea.Cmd {
	return func() tea.Msg {
		repoPath := filepath.Join(m.config.CodebasesPath, repoName)

		cmd := exec.Command("git", "-C", repoPath, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return updateFinishedMsg{err: fmt.Errorf("git pull failed: %w", err)}
		}
		return updateFinishedMsg{err: nil}
	}
}

func (m model) removeRepo(repoName string) tea.Cmd {
	return func() tea.Msg {
		repoPath := filepath.Join(m.config.CodebasesPath, repoName)

		if err := os.RemoveAll(repoPath); err != nil {
			return removeFinishedMsg{err: fmt.Errorf("failed to remove directory: %w", err)}
		}
		return removeFinishedMsg{err: nil}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		// Global keybindings
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.state {
		case stateList:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "c":
				m.state = stateClone
				m.textInput = textinput.New()
				m.textInput.Placeholder = "https://github.com/user/repo"
				m.textInput.Focus()
				m.textInput.Width = 50
			case "u":
				// If no item selected, go to multi-select update
				if m.list.SelectedItem() == nil {
					return initialModel(stateUpdateSelection), nil
				}
				// Otherwise, update single selected item
				m.state = stateUpdate
				m.statusMsg = fmt.Sprintf("Updating %s...", m.list.SelectedItem().(item).Title())
				return m, m.updateRepo(m.list.SelectedItem().(item).Title())
			case "r":
				// If no item selected, go to multi-select remove
				if m.list.SelectedItem() == nil {
					return initialModel(stateRemoveSelection), nil
				}
				// Otherwise, confirm single selected item removal
				m.state = stateRemoveConfirm
			}
		case stateClone:
			switch msg.String() {
			case "enter":
				m.state = stateCloning
				m.statusMsg = fmt.Sprintf("Cloning %s...", m.textInput.Value())
				return m, m.cloneRepo(m.textInput.Value())
			case "esc":
				m.state = stateList
			}
		case stateRemoveConfirm:
			switch msg.String() {
			case "y", "Y":
				m.state = stateUpdate // Use same spinner as update
				m.statusMsg = fmt.Sprintf("Removing %s...", m.list.SelectedItem().(item).Title())
				return m, m.removeRepo(m.list.SelectedItem().(item).Title())
			default: // Any other key cancels
				m.state = stateList
			}
		case stateSetup:
			if msg.String() == "enter" {
				path := m.textInput.Value()
				if path == "" {
					path = m.textInput.Placeholder
				}
				m.config.CodebasesPath = os.ExpandEnv(path)
				if _, err := os.Stat(m.config.CodebasesPath); os.IsNotExist(err) {
					if err := os.MkdirAll(m.config.CodebasesPath, 0755); err != nil {
						m.err = fmt.Errorf("could not create directory: %w", err)
						return m, nil
					}
				}
				if err := m.config.Save(); err != nil {
					m.err = fmt.Errorf("could not save config: %w", err)
					return m, nil
				}
				return initialModel(stateList), nil // Re-run initialModel to load repos
			}
		case stateUpdateSelection, stateRemoveSelection:
			switch msg.String() {
			case " ": // Space key to toggle selection
					index := m.list.Cursor()
					if _, ok := m.selected[index]; ok {
						delete(m.selected, index)
					} else {
						m.selected[index] = struct{}{}
					}
			case "enter":
				// Process selected items
				var selectedRepoNames []string
				for idx := range m.selected {
					selectedRepoNames = append(selectedRepoNames, m.list.Items()[idx].(item).Title())
				}

				if len(selectedRepoNames) == 0 {
					m.state = stateList // Go back if nothing selected
					return m, nil
				}

				// Dispatch commands for each selected repo
				cmds := make([]tea.Cmd, len(selectedRepoNames))
				for i, repoName := range selectedRepoNames {
					if m.state == stateUpdateSelection {
						cmds[i] = m.updateRepo(repoName)
					} else if m.state == stateRemoveSelection {
						cmds[i] = m.removeRepo(repoName)
					}
				}
				m.state = stateUpdate // Use update state for batch operations
				m.statusMsg = fmt.Sprintf("Processing %d repositories...", len(selectedRepoNames))
				return m, tea.Batch(cmds...)
			case "esc":
				m.state = stateList
				return m, nil
			}
		}

	case cloneFinishedMsg, updateFinishedMsg, removeFinishedMsg:
		switch msg := msg.(type) {
		case cloneFinishedMsg:
			if msg.err != nil {
				m.err = msg.err
			}
		case updateFinishedMsg:
			if msg.err != nil {
				m.err = msg.err
			}
		case removeFinishedMsg:
			if msg.err != nil {
				m.err = msg.err
			}
		}
		// Refresh the model
		refreshedModel := initialModel(stateList)
		refreshedModel.err = m.err // Persist error if there was one
		return refreshedModel, nil

	case error:
		m.err = msg
		return m, nil
	}

	switch m.state {
	case stateSetup, stateClone:
		m.textInput, cmd = m.textInput.Update(msg)
	case stateCloning, stateUpdate:
		m.spinner, cmd = m.spinner.Update(msg)
	case stateList, stateUpdateSelection, stateRemoveSelection:
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\nPress any key to return to the list.", m.err)
	}

	var s string
	switch m.state {
	case stateSetup:
		s = fmt.Sprintf("\n%s\n\nPlease enter the path for your codebases directory.\n\n%s\n\n%s", titleStyle.Render("Welcome"), m.textInput.View(), helpStyle.Render("Enter to save"))
	case stateClone:
		s = fmt.Sprintf("\n%s\n\nEnter the repository URL to clone.\n\n%s\n\n%s", titleStyle.Render("Clone"), m.textInput.View(), helpStyle.Render("Enter: clone, Esc: cancel"))
	case stateCloning, stateUpdate:
		s = fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.statusMsg)
	case stateRemoveConfirm:
		selected := m.list.SelectedItem().(item).Title()
		s = fmt.Sprintf("\n\n   Are you sure you want to remove %s?\n   This action cannot be undone.\n\n   [y/N]\n\n", titleStyle.Render(selected))
	case stateList, stateUpdateSelection, stateRemoveSelection:
		var listContent string
		if m.state == stateUpdateSelection || m.state == stateRemoveSelection {
			// Manually render list items with checkboxes
			items := m.list.Items()
			for i, listItem := range items {
				checked := " " // not selected
				if _, ok := m.selected[i]; ok {
					checked = "x" // selected
				}
				
				line := fmt.Sprintf("[%s] %s", checked, listItem.(item).Title())
				if i == m.list.Cursor() {
					line = selectedItemStyle.Render(line)
				}
				listContent += line + "\n"
			}
			s = lipgloss.NewStyle().Margin(1, 2).Render(m.list.Title + "\n" + listContent)
		} else {
			s = lipgloss.NewStyle().Margin(1, 2).Render(m.list.View())
		}
		s += "\n" + m.helpView()
	default:
		s = "Unknown state."
	}
	return s
}

func (m model) helpView() string {
	switch m.state {
	case stateUpdateSelection:
		return helpStyle.Render("  Space: toggle selection | Enter: confirm | Esc: cancel")
	case stateRemoveSelection:
		return helpStyle.Render("  Space: toggle selection | Enter: confirm | Esc: cancel")
	default:
		return helpStyle.Render("  ↑/↓: navigate | c: clone | u: update | r: remove | q: quit")
	}
}

// --- Helper Functions ---

func scanForRepos(root string) ([]string, error) {
	var repos []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			relPath, _ := filepath.Rel(root, filepath.Dir(path))
			repos = append(repos, relPath)
			return filepath.SkipDir
		}
		return nil
	})
	return repos, err
}

func getClonePath(url string) string {
	path := strings.TrimPrefix(url, "https://")
	path = strings.TrimPrefix(path, "http://")
	path = strings.TrimPrefix(path, "git@")
	if strings.Contains(path, ":") {
		path = strings.Replace(path, ":", "/", 1)
	}
	path = strings.TrimSuffix(path, ".git")
	return path
}

// --- Non-interactive command handlers ---

func handleListCommand(cfg config.Config) {
	repos, err := scanForRepos(cfg.CodebasesPath)
	if err != nil {
		fmt.Printf("Error scanning for repos: %v\n", err)
		os.Exit(1)
	}
	for _, repo := range repos {
		fmt.Println(repo)
	}
}

func handleCloneCommand(cfg config.Config, url string) {
	clonePath := getClonePath(url)
	dest := filepath.Join(cfg.CodebasesPath, clonePath)

	fmt.Printf("Cloning %s into %s...\n", url, dest)

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("git", "clone", url, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error cloning repo: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Clone complete.")
}

func handleUpdateCommand(cfg config.Config, repoName string) {
	repoPath := filepath.Join(cfg.CodebasesPath, repoName)
	fmt.Printf("Updating %s...\n", repoName)
	cmd := exec.Command("git", "-C", repoPath, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error updating repo: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Update complete.")
}

func handleRemoveCommand(cfg config.Config, args []string) {
	repoName := args[0]
	repoPath := filepath.Join(cfg.CodebasesPath, repoName)

	force := len(args) > 1 && args[1] == "--force"

	if !force {
		fmt.Printf("Are you sure you want to remove %s? [y/N] ", repoName)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(input)) != "y" {
			fmt.Println("Remove cancelled.")
			os.Exit(0)
		}
	}

	fmt.Printf("Removing %s...\n", repoName)
	if err := os.RemoveAll(repoPath); err != nil {
		fmt.Printf("Error removing repo: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Repository removed.")
}

func launchTUI(initialAppState state) {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	p := tea.NewProgram(initialModel(initialAppState), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func main() {
	// Check for interactive flags first
	if len(os.Args) > 1 && (os.Args[1] == "-i" || os.Args[1] == "--interactive") {
		launchTUI(stateList) // Launch TUI in default list mode
		return
	}

	// Handle non-interactive commands or specific TUI modes
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		args := os.Args[2:] // Remaining arguments

		// Commands that don't require config to be set (e.g., help, version, or setup)
		// For now, all commands require config, so load it here.
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
		if cfg.CodebasesPath == "" && cmd != "list" && cmd != "update" && cmd != "remove" {
			// If config is not set, and it's not a command that can launch TUI for setup,
			// then we can't proceed with non-interactive commands.
			fmt.Println("Error: VCS_CODEBASES path not set. Please run 'get-repo' interactively to configure.")
			os.Exit(1)
		}

		switch cmd {
		case "list":
			handleListCommand(cfg)
		case "update":
			if len(args) == 0 { // get-repo update (no args) -> interactive update
				launchTUI(stateUpdateSelection)
			} else { // get-repo update <repo-name> -> non-interactive update
				handleUpdateCommand(cfg, args[0])
			}
		case "remove":
			if len(args) == 0 { // get-repo remove (no args) -> interactive remove
				launchTUI(stateRemoveSelection)
			} else { // get-repo remove <repo-name> [--force] -> non-interactive remove
				handleRemoveCommand(cfg, args)
			}
		default: // Assume it's a URL for cloning
			if strings.HasPrefix(cmd, "http") || strings.HasPrefix(cmd, "git@") {
				handleCloneCommand(cfg, cmd) // cmd is the URL
			} else {
				fmt.Printf("Unknown command or invalid URL: %s\n", cmd)
				fmt.Println("Usage: get-repo [url] | [list|update|remove] | [-i|--interactive]")
				os.Exit(1)
			}
		}
		return // Exit after handling non-interactive command or launching specific TUI
	}

	// Default to interactive mode if no arguments
	launchTUI(stateList)
}