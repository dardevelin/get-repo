package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PathInput provides a text input with path validation and tab completion
type PathInput struct {
	textInput     textinput.Model
	lastPath      string
	pathExists    bool
	isDirectory   bool
	completions   []string
	showingCompletion bool
}

// NewPathInput creates a new path input with validation
func NewPathInput() PathInput {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60
	
	return PathInput{
		textInput: ti,
	}
}

// Update handles path input updates including validation and tab completion
func (p PathInput) Update(msg tea.Msg) (PathInput, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Handle tab completion
			currentPath := p.textInput.Value()
			if currentPath != "" {
				completion := p.getTabCompletion(currentPath)
				if completion != "" {
					p.textInput.SetValue(completion)
					p.showingCompletion = true
				}
			}
			return p, nil
		default:
			p.showingCompletion = false
		}
	}
	
	// Update the text input
	p.textInput, cmd = p.textInput.Update(msg)
	
	// Validate path if it changed
	currentPath := p.textInput.Value()
	if currentPath != p.lastPath {
		p.lastPath = currentPath
		p.validatePath(currentPath)
	}
	
	return p, cmd
}

// validatePath checks if the current path exists and updates validation state
func (p *PathInput) validatePath(path string) {
	if path == "" {
		p.pathExists = false
		p.isDirectory = false
		return
	}
	
	// Expand environment variables
	expandedPath := os.ExpandEnv(path)
	
	// Check if path exists
	info, err := os.Stat(expandedPath)
	if err != nil {
		p.pathExists = false
		p.isDirectory = false
		
		// Check if parent directory exists (for new directories)
		parent := filepath.Dir(expandedPath)
		if parentInfo, parentErr := os.Stat(parent); parentErr == nil && parentInfo.IsDir() {
			// Parent exists, this could be a valid new directory
			p.pathExists = true // We'll consider this "valid" for styling
		}
		return
	}
	
	p.pathExists = true
	p.isDirectory = info.IsDir()
}

// getTabCompletion provides tab completion for the current path
func (p *PathInput) getTabCompletion(currentPath string) string {
	if currentPath == "" {
		return ""
	}
	
	expandedPath := os.ExpandEnv(currentPath)
	
	// If the path ends with a separator, list contents of that directory
	if strings.HasSuffix(expandedPath, string(filepath.Separator)) {
		return p.completeDirectory(expandedPath, "")
	}
	
	// Otherwise, complete the current path segment
	dir := filepath.Dir(expandedPath)
	base := filepath.Base(expandedPath)
	
	return p.completeDirectory(dir, base)
}

// completeDirectory finds the first matching completion in a directory
func (p *PathInput) completeDirectory(dir, prefix string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	
	var matches []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
			fullPath := filepath.Join(dir, name)
			if entry.IsDir() {
				fullPath += string(filepath.Separator)
			}
			matches = append(matches, fullPath)
		}
	}
	
	if len(matches) > 0 {
		sort.Strings(matches)
		return matches[0]
	}
	
	return ""
}

// View renders the path input with validation styling
func (p PathInput) View() string {
	// Get the current value for styling
	currentPath := p.textInput.Value()
	
	// Style the input based on validation state
	var styledInput string
	if currentPath == "" {
		// No input yet
		styledInput = p.textInput.View()
	} else if p.pathExists {
		if p.isDirectory {
			// Valid directory - green
			styledInput = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Render(p.textInput.View())
		} else {
			// Valid file - blue (though we mainly expect directories)
			styledInput = lipgloss.NewStyle().
				Foreground(lipgloss.Color("12")).
				Render(p.textInput.View())
		}
	} else {
		// Invalid/non-existent path - red
		styledInput = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render(p.textInput.View())
	}
	
	// Add validation indicator
	var indicator string
	if currentPath != "" {
		if p.pathExists {
			if p.isDirectory {
				indicator = SuccessStyle.Render(" ✓ Directory exists")
			} else {
				indicator = lipgloss.NewStyle().
					Foreground(lipgloss.Color("12")).
					Render(" ✓ File exists")
			}
		} else {
			// Check if it could be a new directory
			expandedPath := os.ExpandEnv(currentPath)
			parent := filepath.Dir(expandedPath)
			if parentInfo, err := os.Stat(parent); err == nil && parentInfo.IsDir() {
				indicator = lipgloss.NewStyle().
					Foreground(lipgloss.Color("11")).
					Render(" ⚠ Will be created")
			} else {
				indicator = ErrorStyle.Render(" ✗ Path invalid")
			}
		}
	}
	
	result := styledInput
	if indicator != "" {
		result += "\n" + indicator
	}
	
	// Add tab completion hint
	if currentPath != "" && !p.showingCompletion {
		result += "\n" + HelpStyle.Render("Press Tab for completion")
	}
	
	return result
}

// Value returns the current input value
func (p PathInput) Value() string {
	return p.textInput.Value()
}

// SetValue sets the input value
func (p *PathInput) SetValue(value string) {
	p.textInput.SetValue(value)
	p.validatePath(value)
}

// SetPlaceholder sets the placeholder text
func (p *PathInput) SetPlaceholder(placeholder string) {
	p.textInput.Placeholder = placeholder
}

// Focus focuses the input
func (p *PathInput) Focus() tea.Cmd {
	return p.textInput.Focus()
}

// Blur removes focus from the input
func (p PathInput) Blur() {
	p.textInput.Blur()
}

// Focused returns whether the input is focused
func (p PathInput) Focused() bool {
	return p.textInput.Focused()
}