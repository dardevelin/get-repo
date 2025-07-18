package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"get-repo/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SetupWizard manages the first-run setup flow
type SetupWizard struct {
	step              SetupStep
	configLocation    string
	codebasesPath     string
	useCustomLocation bool
	shellChoice       string
	textInput         textinput.Model
	pathInput         PathInput
	selectedIndex     int
	choices           []string
	fileBrowser       FileBrowser
	browserMode       BrowserMode
}

type BrowserMode int

const (
	BrowserModeSelect BrowserMode = iota
	BrowserModeType
)

type SetupStep int

const (
	StepWelcome SetupStep = iota
	StepConfigLocation
	StepCustomConfigPath
	StepCustomConfigBrowser
	StepCodebasesPath
	StepCodebasesBrowser
	StepShellIntegration
	StepReview
	StepComplete
)

// NewSetupWizard creates a new setup wizard
func NewSetupWizard() SetupWizard {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60

	return SetupWizard{
		step:        StepWelcome,
		textInput:   ti,
		pathInput:   NewPathInput(),
		browserMode: BrowserModeSelect,
		choices: []string{
			"Use default location (~/.config/get-repo)",
			"Choose custom location",
		},
	}
}

func (s SetupWizard) Update(msg tea.Msg) (SetupWizard, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Forward window size messages to file browser when active
		if s.step == StepCustomConfigBrowser || s.step == StepCodebasesBrowser {
			s.fileBrowser, cmd = s.fileBrowser.Update(msg)
			return s, cmd
		}
		
	case tea.KeyMsg:
		// Global back navigation (except for welcome, complete, and browser steps)
		if msg.String() == "esc" && s.step != StepWelcome && s.step != StepComplete && 
		   s.step != StepCustomConfigBrowser && s.step != StepCodebasesBrowser {
			return s.goBack(), nil
		}
		
		switch s.step {
		case StepWelcome:
			if msg.String() == "enter" {
				s.step = StepConfigLocation
			}

		case StepConfigLocation:
			switch msg.String() {
			case "up", "k":
				s.selectedIndex = 0
			case "down", "j":
				s.selectedIndex = 1
			case "enter":
				s.useCustomLocation = s.selectedIndex == 1
				if s.useCustomLocation {
					s.step = StepCustomConfigPath
					s.choices = []string{
						"Browse for directory",
						"Type path manually",
					}
					s.selectedIndex = 0
				} else {
					// Use default location
					defaultPath, _ := os.UserConfigDir()
					s.configLocation = filepath.Join(defaultPath, "get-repo", "config.json")
					s.step = StepCodebasesPath
					s.choices = []string{
						"Use default location (~/.../dev/vcs-codebases)",
						"Browse for directory", 
						"Type path manually",
					}
					s.selectedIndex = 0
				}
			}

		case StepCustomConfigPath:
			switch msg.String() {
			case "up", "k":
				s.selectedIndex = 0
			case "down", "j":
				s.selectedIndex = 1
			case "enter":
				if s.selectedIndex == 0 {
					// Browse for directory
					s.step = StepCustomConfigBrowser
					s.fileBrowser = NewFileBrowser(os.Getenv("HOME"), true)
				} else {
					// Type manually
					s.browserMode = BrowserModeType
					s.pathInput.SetPlaceholder(filepath.Join(os.Getenv("HOME"), ".get-repo"))
					s.pathInput.SetValue("")
					s.pathInput.Focus()
				}
			}
			
			if s.browserMode == BrowserModeType && msg.String() == "enter" {
				path := s.pathInput.Value()
				if path == "" {
					path = filepath.Join(os.Getenv("HOME"), ".get-repo") // Default
				}
				s.configLocation = filepath.Join(os.ExpandEnv(path), "config.json")
				s.step = StepCodebasesPath
				s.browserMode = BrowserModeSelect
				s.choices = []string{
					"Use default location (~/.../dev/vcs-codebases)",
					"Browse for directory", 
					"Type path manually",
				}
				s.selectedIndex = 0
			}

		case StepCustomConfigBrowser:
			switch msg.String() {
			case "enter":
				if selectedItem := s.fileBrowser.GetSelectedItem(); selectedItem != nil {
					if selectedItem.name == "📍 Select this directory" {
						// User selected current directory
						s.configLocation = filepath.Join(s.fileBrowser.GetCurrentPath(), "config.json")
						s.step = StepCodebasesPath
						s.choices = []string{
							"Use default location (~/.../dev/vcs-codebases)",
							"Browse for directory", 
							"Type path manually",
						}
						s.selectedIndex = 0
					} else {
						// Navigate into directory or up
						s.fileBrowser, cmd = s.fileBrowser.Update(msg)
					}
				}
			case " ": // Space key also works to select current directory
				s.configLocation = filepath.Join(s.fileBrowser.GetCurrentPath(), "config.json")
				s.step = StepCodebasesPath
				s.choices = []string{
					"Use default location (~/.../dev/vcs-codebases)",
					"Browse for directory", 
					"Type path manually",
				}
				s.selectedIndex = 0
			case "esc":
				s.step = StepCustomConfigPath
				s.browserMode = BrowserModeSelect
			default:
				s.fileBrowser, cmd = s.fileBrowser.Update(msg)
			}

		case StepCodebasesPath:
			switch msg.String() {
			case "up", "k":
				s.selectedIndex--
				if s.selectedIndex < 0 {
					s.selectedIndex = len(s.choices) - 1
				}
			case "down", "j":
				s.selectedIndex = (s.selectedIndex + 1) % len(s.choices)
			case "enter":
				if s.selectedIndex == 0 {
					// Use default location
					defaultPath := filepath.Join(os.Getenv("HOME"), "dev", "vcs-codebases")
					s.codebasesPath = defaultPath
					
					// Check if we need shell integration
					if s.useCustomLocation {
						s.step = StepShellIntegration
						s.choices = []string{"zsh (~/.zshrc)", "bash (~/.bashrc)", "fish (~/.config/fish/config.fish)", "Skip"}
						s.selectedIndex = 0
					} else {
						s.step = StepReview
					}
				} else if s.selectedIndex == 1 {
					// Browse for directory
					s.step = StepCodebasesBrowser
					s.fileBrowser = NewFileBrowser(os.Getenv("HOME"), true)
				} else {
					// Type manually
					s.browserMode = BrowserModeType
					s.pathInput.SetPlaceholder(filepath.Join(os.Getenv("HOME"), "dev", "vcs-codebases"))
					s.pathInput.SetValue("")
					s.pathInput.Focus()
				}
			}
			
			if s.browserMode == BrowserModeType && msg.String() == "enter" {
				path := s.pathInput.Value()
				if path == "" {
					path = filepath.Join(os.Getenv("HOME"), "dev", "vcs-codebases") // Default
				}
				s.codebasesPath = os.ExpandEnv(path)
				
				// Check if we need shell integration
				if s.useCustomLocation {
					s.step = StepShellIntegration
					s.choices = []string{"zsh (~/.zshrc)", "bash (~/.bashrc)", "fish (~/.config/fish/config.fish)", "Skip"}
					s.selectedIndex = 0
				} else {
					s.step = StepReview
				}
				s.browserMode = BrowserModeSelect
			}

		case StepCodebasesBrowser:
			switch msg.String() {
			case "enter":
				if selectedItem := s.fileBrowser.GetSelectedItem(); selectedItem != nil {
					if selectedItem.name == "📍 Select this directory" {
						// User selected current directory
						s.codebasesPath = s.fileBrowser.GetCurrentPath()
						
						// Check if we need shell integration
						if s.useCustomLocation {
							s.step = StepShellIntegration
							s.choices = []string{"zsh (~/.zshrc)", "bash (~/.bashrc)", "fish (~/.config/fish/config.fish)", "Skip"}
							s.selectedIndex = 0
						} else {
							s.step = StepReview
						}
					} else {
						// Navigate into directory or up
						s.fileBrowser, cmd = s.fileBrowser.Update(msg)
					}
				}
			case " ": // Space key also works to select current directory
				s.codebasesPath = s.fileBrowser.GetCurrentPath()
				
				// Check if we need shell integration
				if s.useCustomLocation {
					s.step = StepShellIntegration
					s.choices = []string{"zsh (~/.zshrc)", "bash (~/.bashrc)", "fish (~/.config/fish/config.fish)", "Skip"}
					s.selectedIndex = 0
				} else {
					s.step = StepReview
				}
			case "esc":
				s.step = StepCodebasesPath
				s.browserMode = BrowserModeSelect
			default:
				s.fileBrowser, cmd = s.fileBrowser.Update(msg)
			}

		case StepShellIntegration:
			switch msg.String() {
			case "up", "k":
				s.selectedIndex--
				if s.selectedIndex < 0 {
					s.selectedIndex = len(s.choices) - 1
				}
			case "down", "j":
				s.selectedIndex = (s.selectedIndex + 1) % len(s.choices)
			case "enter":
				switch s.selectedIndex {
				case 0:
					s.shellChoice = "zsh"
				case 1:
					s.shellChoice = "bash"
				case 2:
					s.shellChoice = "fish"
				case 3:
					s.shellChoice = "skip"
				}
				s.step = StepReview
			}

		case StepReview:
			switch msg.String() {
			case "enter", "y", "Y":
				s.step = StepComplete
			case "e", "E":
				// Edit - go back to start (could add more granular editing later)
				s.step = StepConfigLocation
			}
		}
	}

	// Update path input only when in typing mode
	if s.browserMode == BrowserModeType {
		s.pathInput, cmd = s.pathInput.Update(msg)
	}

	return s, cmd
}

// goBack handles backward navigation through the wizard
func (s SetupWizard) goBack() SetupWizard {
	switch s.step {
	case StepConfigLocation:
		s.step = StepWelcome
	case StepCustomConfigPath:
		s.step = StepConfigLocation
	case StepCustomConfigBrowser:
		s.step = StepCustomConfigPath
		s.browserMode = BrowserModeSelect
	case StepCodebasesPath:
		if s.useCustomLocation {
			s.step = StepCustomConfigPath
		} else {
			s.step = StepConfigLocation
		}
	case StepCodebasesBrowser:
		s.step = StepCodebasesPath
		s.browserMode = BrowserModeSelect
	case StepShellIntegration:
		s.step = StepCodebasesPath
	case StepReview:
		if s.useCustomLocation {
			s.step = StepShellIntegration
		} else {
			s.step = StepCodebasesPath
		}
	}
	
	// Reset browser mode when going back
	if s.step != StepCustomConfigBrowser && s.step != StepCodebasesBrowser {
		s.browserMode = BrowserModeSelect
	}
	
	return s
}

func (s SetupWizard) View() string {
	switch s.step {
	case StepWelcome:
		return fmt.Sprintf(`
%s

Welcome! Let's set up get-repo for the first time.

This wizard will help you:
• Choose where to store your configuration
• Set up your repositories directory
• Configure shell integration (if needed)

Press %s to continue.`,
			TitleStyle.Render("Get-Repo Setup"),
			HelpStyle.Render("Enter to continue • Ctrl+C to quit"))

	case StepConfigLocation:
		// Show the default path
		defaultPath, _ := os.UserConfigDir()
		defaultConfigPath := filepath.Join(defaultPath, "get-repo", "config.json")
		
		choices := ""
		for i, choice := range s.choices {
			cursor := "  "
			if i == s.selectedIndex {
				cursor = SelectedItemStyle.Render("→ ")
				choice = SelectedItemStyle.Render(choice)
			}
			choices += cursor + choice + "\n"
		}

		return fmt.Sprintf(`
%s

Where would you like to store the configuration?

Default: %s

%s
%s`,
			TitleStyle.Render("Configuration Location"),
			HelpStyle.Render(defaultConfigPath),
			choices,
			HelpStyle.Render("↑/↓ to select • Enter to confirm • Esc to go back"))

	case StepCustomConfigPath:
		if s.browserMode == BrowserModeType {
			return fmt.Sprintf(`
%s

Enter the directory path for your configuration:

%s

%s`,
				TitleStyle.Render("Custom Configuration Directory"),
				s.pathInput.View(),
				HelpStyle.Render("Enter to confirm • Esc to go back • Tab for completion"))
		}
		
		choices := ""
		for i, choice := range s.choices {
			cursor := "  "
			if i == s.selectedIndex {
				cursor = SelectedItemStyle.Render("→ ")
				choice = SelectedItemStyle.Render(choice)
			}
			choices += cursor + choice + "\n"
		}

		return fmt.Sprintf(`
%s

How would you like to choose your configuration directory?

%s
%s`,
			TitleStyle.Render("Custom Configuration Directory"),
			choices,
			HelpStyle.Render("↑/↓ to select • Enter to confirm • Esc to go back"))

	case StepCustomConfigBrowser:
		return fmt.Sprintf(`
%s

Select a directory for your configuration file:
(config.json will be created in the selected directory)

%s

%s`,
			TitleStyle.Render("Browse Configuration Directory"),
			s.fileBrowser.View(),
			HelpStyle.Render("Enter: navigate/select • Space: select current • h: hidden files • Esc: back"))

	case StepCodebasesPath:
		if s.browserMode == BrowserModeType {
			return fmt.Sprintf(`
%s

Enter the path where you keep your git repositories:

%s

%s`,
				TitleStyle.Render("Repositories Directory"),
				s.pathInput.View(),
				HelpStyle.Render("Enter to confirm • Esc to go back • Tab for completion"))
		}
		
		// Show the default path
		defaultRepoPath := filepath.Join(os.Getenv("HOME"), "dev", "vcs-codebases")
		
		choices := ""
		for i, choice := range s.choices {
			cursor := "  "
			if i == s.selectedIndex {
				cursor = SelectedItemStyle.Render("→ ")
				choice = SelectedItemStyle.Render(choice)
			}
			choices += cursor + choice + "\n"
		}

		return fmt.Sprintf(`
%s

Where do you keep your git repositories?

Default: %s

%s
%s`,
			TitleStyle.Render("Repositories Directory"),
			HelpStyle.Render(defaultRepoPath),
			choices,
			HelpStyle.Render("↑/↓ to select • Enter to confirm • Esc to go back"))

	case StepCodebasesBrowser:
		return fmt.Sprintf(`
%s

Select a directory for your git repositories:

%s

%s`,
			TitleStyle.Render("Browse Repositories Directory"),
			s.fileBrowser.View(),
			HelpStyle.Render("Enter: navigate/select • Space: select current • h: hidden files • Esc: back"))

	case StepShellIntegration:
		choices := ""
		for i, choice := range s.choices {
			cursor := "  "
			if i == s.selectedIndex {
				cursor = SelectedItemStyle.Render("→ ")
				choice = SelectedItemStyle.Render(choice)
			}
			choices += cursor + choice + "\n"
		}

		return fmt.Sprintf(`
%s

Since you're using a custom config location, we need to set the
GET_REPO_CONFIG environment variable.

Which shell configuration should we update?

%s

%s`,
			TitleStyle.Render("Shell Integration"),
			choices,
			HelpStyle.Render("↑/↓ to select • Enter to confirm • Esc to go back"))

	case StepReview:
		configPath := s.configLocation
		if configPath == "" {
			defaultPath, _ := os.UserConfigDir()
			configPath = filepath.Join(defaultPath, "get-repo", "config.json")
		}
		
		summary := fmt.Sprintf(`
%s

Please review your configuration:

• Configuration file: %s
• Repositories directory: %s`,
			TitleStyle.Render("Review Configuration"),
			configPath,
			s.codebasesPath)

		if s.useCustomLocation && s.shellChoice != "" && s.shellChoice != "skip" {
			summary += fmt.Sprintf("\n• Shell integration: %s", s.shellChoice)
		}

		summary += "\n\nIs this correct?"
		summary += "\n\n" + HelpStyle.Render("Enter/Y: Apply configuration • E: Edit • Esc: Go back")

		return summary

	case StepComplete:
		summary := fmt.Sprintf(`
%s

Setup complete! Here's what we configured:

• Config location: %s
• Repositories: %s`,
			SuccessStyle.Render("✓ All Done!"),
			s.configLocation,
			s.codebasesPath)

		if s.shellChoice != "" && s.shellChoice != "skip" {
			summary += fmt.Sprintf("\n• Shell integration: %s", s.shellChoice)
		}

		summary += "\n\nPress any key to start using get-repo!"

		return summary

	default:
		return "Unknown setup step"
	}
}

// Apply saves the configuration and sets up shell integration
func (s SetupWizard) Apply() error {
	// Create the configuration
	cfg := config.Config{
		CodebasesPath: s.codebasesPath,
	}

	// Create codebases directory if it doesn't exist
	if err := os.MkdirAll(s.codebasesPath, 0755); err != nil {
		return fmt.Errorf("failed to create codebases directory: %w", err)
	}

	// Save configuration
	if err := cfg.SaveTo(s.configLocation); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Set up shell integration if needed
	if s.useCustomLocation && s.shellChoice != "" && s.shellChoice != "skip" {
		if err := s.setupShellIntegration(); err != nil {
			// Don't fail the whole setup if shell integration fails
			fmt.Fprintf(os.Stderr, "Warning: Failed to set up shell integration: %v\n", err)
		}
	}

	return nil
}

func (s SetupWizard) setupShellIntegration() error {
	exportLine := fmt.Sprintf("export %s=\"%s\"", config.EnvConfigPath, s.configLocation)
	comment := "# get-repo configuration"

	var rcFile string
	switch s.shellChoice {
	case "zsh":
		rcFile = filepath.Join(os.Getenv("HOME"), ".zshrc")
	case "bash":
		rcFile = filepath.Join(os.Getenv("HOME"), ".bashrc")
	case "fish":
		rcFile = filepath.Join(os.Getenv("HOME"), ".config", "fish", "config.fish")
		exportLine = fmt.Sprintf("set -x %s \"%s\"", config.EnvConfigPath, s.configLocation)
	default:
		return nil
	}

	// Read existing file
	content, err := os.ReadFile(rcFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", rcFile, err)
	}

	// Check if already configured
	if strings.Contains(string(content), config.EnvConfigPath) {
		return nil // Already configured
	}

	// Append configuration
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", rcFile, err)
	}
	defer f.Close()

	// Add newline if file doesn't end with one
	if len(content) > 0 && content[len(content)-1] != '\n' {
		f.WriteString("\n")
	}

	// Write configuration
	_, err = f.WriteString(fmt.Sprintf("\n%s\n%s\n", comment, exportLine))
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", rcFile, err)
	}

	return nil
}