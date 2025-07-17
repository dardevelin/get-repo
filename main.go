package main

import (
	"fmt"
	"io/fs"
	"log"
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
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#5D5D5D")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type model struct {
	state     state
	config    config.Config
	list      list.Model
	textInput textinput.Model
	spinner   spinner.Model
	statusMsg string
	err       error
}

func initialModel() model {
	cfg, err := config.Load()
	if err != nil {
		return model{err: err}
	}

	if cfg.CodebasesPath == "" {
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

	repos, err := scanForRepos(cfg.CodebasesPath)
	if err != nil {
		return model{err: err}
	}

	items := make([]list.Item, len(repos))
	for i, repo := range repos {
		items[i] = item(repo)
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Your Repositories"
	l.SetShowHelp(false)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		state:   stateList,
		config:  cfg,
		list:    l,
		spinner: s,
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
func (m model) cloneRepo() tea.Cmd {
	return func() tea.Msg {
		url := m.textInput.Value()
		clonePath := getClonePath(url)
		dest := filepath.Join(m.config.CodebasesPath, clonePath)

		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return cloneFinishedMsg{err: fmt.Errorf("failed to create directory: %w", err)}
		}

		cmd := exec.Command("git", "clone", url, dest)
		if err := cmd.Run(); err != nil {
			return cloneFinishedMsg{err: fmt.Errorf("git clone failed: %w", err)}
		}
		return cloneFinishedMsg{err: nil}
	}
}

func (m model) updateRepo() tea.Cmd {
	return func() tea.Msg {
		selectedItem := m.list.SelectedItem().(item)
		repoPath := filepath.Join(m.config.CodebasesPath, string(selectedItem))

		cmd := exec.Command("git", "-C", repoPath, "pull")
		if err := cmd.Run(); err != nil {
			return updateFinishedMsg{err: fmt.Errorf("git pull failed: %w", err)}
		}
		return updateFinishedMsg{err: nil}
	}
}

func (m model) removeRepo() tea.Cmd {
	return func() tea.Msg {
		selectedItem := m.list.SelectedItem().(item)
		repoPath := filepath.Join(m.config.CodebasesPath, string(selectedItem))

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
				if m.list.SelectedItem() != nil {
					m.state = stateUpdate
					m.statusMsg = fmt.Sprintf("Updating %s...", m.list.SelectedItem().(item).Title())
					return m, m.updateRepo()
				}
			case "r":
				if m.list.SelectedItem() != nil {
					m.state = stateRemoveConfirm
				}
			}
		case stateClone:
			switch msg.String() {
			case "enter":
				m.state = stateCloning
				m.statusMsg = fmt.Sprintf("Cloning %s...", m.textInput.Value())
				return m, m.cloneRepo()
			case "esc":
				m.state = stateList
			}
		case stateRemoveConfirm:
			switch msg.String() {
			case "y", "Y":
				m.state = stateUpdate // Use same spinner as update
				m.statusMsg = fmt.Sprintf("Removing %s...", m.list.SelectedItem().(item).Title())
				return m, m.removeRepo()
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
				return initialModel(), nil
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
		refreshedModel := initialModel()
		refreshedModel.err = m.err // Persist error if there was one
		return refreshedModel, nil

	case error:
		m.err = msg
		return m, nil
	}

	switch m.state {
	case stateSetup, stateClone:
		m.textInput, cmd = m.textInput.Update(msg)
	case stateUpdate, stateCloning:
		m.spinner, cmd = m.spinner.Update(msg)
	default:
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\nPress any key to return to the list.", m.err)
	}

	switch m.state {
	case stateSetup:
		return fmt.Sprintf("\n%s\n\nPlease enter the path for your codebases directory.\n\n%s\n\n%s", titleStyle.Render("Welcome"), m.textInput.View(), helpStyle.Render("Enter to save"))
	case stateClone:
		return fmt.Sprintf("\n%s\n\nEnter the repository URL to clone.\n\n%s\n\n%s", titleStyle.Render("Clone"), m.textInput.View(), helpStyle.Render("Enter: clone, Esc: cancel"))
	case stateCloning, stateUpdate:
		return fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.statusMsg)
	case stateRemoveConfirm:
		selected := m.list.SelectedItem().(item).Title()
		return fmt.Sprintf("\n\n   Are you sure you want to remove %s?\n   This action cannot be undone.\n\n   [y/N]\n\n", titleStyle.Render(selected))
	case stateList:
		return lipgloss.NewStyle().Margin(1, 2).Render(m.list.View()) + "\n" + m.helpView()
	default:
		return "Unknown state."
	}
}

func (m model) helpView() string {
	return helpStyle.Render("  ↑/↓: navigate | c: clone | u: update | r: remove | q: quit")
}

// --- Helper Functions ---

func scanForRepos(root string) ([]string, error) {
	var repos []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			relPath, _ := filepath.rel(root, filepath.Dir(path))
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

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
