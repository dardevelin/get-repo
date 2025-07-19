package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FileBrowser provides a simple file/directory browser
type FileBrowser struct {
	currentPath   string
	items         []list.Item
	list          list.Model
	showHidden    bool
	directoryOnly bool
	terminalWidth int
}

// FileItem represents a file or directory in the browser
type FileItem struct {
	name     string
	path     string
	isDir    bool
	isHidden bool
}

func (f FileItem) Title() string {
	// Special handling for special entries - no icons, just clean text
	if f.name == "📍 Select this directory" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ec9b0")). // Muted cyan like eza
			Bold(true).
			Render("󰄬  Select this directory")
	}

	if f.name == ".." {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")). // Muted gray
			Render("  .. (parent directory)")
	}

	// Get icon and color for the item
	icon, color := getFileIconAndColor(f.name, f.isDir)

	// Apply styling based on file state
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	if f.isDir && !f.isHidden {
		style = style.Bold(true)
	}
	if f.isHidden {
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")). // Muted gray for hidden files
			Italic(true)
	}

	styledName := style.Render(f.name)
	return fmt.Sprintf("%s  %s", icon, styledName)
}

func (f FileItem) Description() string {
	// Remove descriptions entirely for cleaner look
	return ""
}

// getFileIconAndColor returns an appropriate icon and color for files/directories
func getFileIconAndColor(filename string, isDir bool) (string, string) {
	if isDir {
		// Directory icon and color - real eza style (blue from screenshot)
		return "", "#569cd6" // Blue for directories like in eza
	}

	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".go":
		return "", "#569cd6" // Go files - blue like in eza
	case ".js", ".jsx":
		return "", "#dcdcaa" // JavaScript - muted yellow
	case ".ts", ".tsx":
		return "", "#569cd6" // TypeScript - blue
	case ".py":
		return "", "#dcdcaa" // Python - muted yellow
	case ".rs":
		return "", "#ce9178" // Rust - orange/brown
	case ".java":
		return "󰬷", "#ce9178" // Java - orange/brown
	case ".cpp", ".c", ".cc", ".cxx":
		return "", "#569cd6" // C++ - blue
	case ".h", ".hpp":
		return "", "#c586c0" // Header - purple
	case ".md":
		return "", "#d4d4d4" // Markdown - light gray
	case ".txt":
		return "", "#d4d4d4" // Text - light gray
	case ".json":
		return "", "#dcdcaa" // JSON - muted yellow
	case ".yaml", ".yml":
		return "", "#f44747" // YAML - red
	case ".toml":
		return "", "#ce9178" // TOML - orange
	case ".xml":
		return "", "#ce9178" // XML - orange
	case ".html", ".htm":
		return "", "#f44747" // HTML - red
	case ".css":
		return "", "#569cd6" // CSS - blue
	case ".scss", ".sass":
		return "", "#c586c0" // Sass - purple
	case ".sh", ".bash", ".zsh":
		return "", "#4ec9b0" // Shell - cyan/green
	case ".git":
		return "", "#f44747" // Git - red
	case ".gitignore", ".gitmodules":
		return "", "#808080" // Git files - gray
	case ".env", ".env.local", ".env.example":
		return "", "#dcdcaa" // Env - yellow
	case ".docker", ".dockerfile":
		return "", "#569cd6" // Docker - blue
	case ".sql":
		return "", "#569cd6" // SQL - blue
	case ".log":
		return "", "#808080" // Log - gray
	case ".pdf":
		return "", "#f44747" // PDF - red
	case ".zip", ".tar", ".gz", ".7z", ".rar":
		return "", "#ce9178" // Archives - orange
	case ".img", ".iso":
		return "", "#c586c0" // Images - purple
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp":
		return "", "#c586c0" // Images - purple
	case ".mp4", ".avi", ".mkv", ".mov":
		return "", "#ce9178" // Video - orange
	case ".mp3", ".wav", ".flac", ".ogg":
		return "", "#c586c0" // Audio - purple
	default:
		return "", "#d4d4d4" // Default - light gray
	}
}

func (f FileItem) FilterValue() string {
	return f.name
}

// NewFileBrowser creates a new file browser starting at the given path
func NewFileBrowser(startPath string, directoryOnly bool) FileBrowser {
	if startPath == "" {
		startPath = os.Getenv("HOME")
	}

	// Ensure the path exists and is absolute
	absPath, err := filepath.Abs(startPath)
	if err != nil || !pathExists(absPath) {
		absPath = os.Getenv("HOME")
	}

	fb := FileBrowser{
		currentPath:   absPath,
		showHidden:    false,
		directoryOnly: directoryOnly,
		terminalWidth: 80, // Default fallback
	}

	fb.refreshItems()

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false // No descriptions for cleaner look
	delegate.SetHeight(1)            // Single line items
	delegate.SetSpacing(0)

	// Simple, clean selection style
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")). // White text
		Background(lipgloss.Color("#264f78")). // Darker blue background
		Padding(0, 1)

	// Remove all other styling for minimal look
	delegate.Styles.NormalTitle = lipgloss.NewStyle()
	delegate.Styles.DimmedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	delegate.Styles.FilterMatch = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ec9b0"))

	fb.list = list.New(fb.items, delegate, 80, 20) // Start with reasonable size
	fb.list.Title = ""                             // No title, we'll handle it in the view
	fb.list.SetShowHelp(false)
	fb.list.SetShowStatusBar(false) // Hide status bar
	fb.list.SetShowTitle(false)     // Don't show the title in the list
	fb.list.SetFilteringEnabled(true)
	fb.list.DisableQuitKeybindings()

	return fb
}

// Update handles file browser updates
func (fb FileBrowser) Update(msg tea.Msg) (FileBrowser, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Store terminal width for status bar calculations
		fb.terminalWidth = msg.Width

		// Calculate available space more precisely
		// Account for: border (2+2), padding (1+1), status bar (1), help text (1), margins (2)
		availableWidth := msg.Width - 8    // Border + padding + margins
		availableHeight := msg.Height - 10 // Status bar + help text + border + padding + margins

		// Ensure minimum usable dimensions
		if availableWidth < 40 {
			availableWidth = 40
		}
		if availableHeight < 6 {
			availableHeight = 6
		}

		fb.list.SetSize(availableWidth, availableHeight)

		return fb, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if selectedItem := fb.list.SelectedItem(); selectedItem != nil {
				fileItem := selectedItem.(FileItem)
				if fileItem.isDir {
					if fileItem.name == "📍 Select this directory" {
						// This is handled by the parent component
						return fb, nil
					} else if fileItem.name == ".." {
						// Go up one directory
						newPath := filepath.Dir(fb.currentPath)
						fb.currentPath = newPath
						fb.refreshItems()
						fb.list.SetItems(fb.items)
						fb.list.ResetSelected()
						return fb, nil
					} else {
						// Navigate into directory
						fb.currentPath = fileItem.path
						fb.refreshItems()
						fb.list.SetItems(fb.items)
						fb.list.ResetSelected()
						return fb, nil
					}
				}
			}

		case "h", "ctrl+h":
			// Toggle hidden files
			fb.showHidden = !fb.showHidden
			fb.refreshItems()
			fb.list.SetItems(fb.items)
			return fb, nil

		case "ctrl+l":
			// Refresh current directory
			fb.refreshItems()
			fb.list.SetItems(fb.items)
			return fb, nil
		}
	}

	fb.list, cmd = fb.list.Update(msg)
	return fb, cmd
}

// View renders the file browser with clean styling
func (fb FileBrowser) View() string {
	// Simple header with current path
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#569cd6")).
		Bold(true).
		MarginBottom(1).
		Render(fmt.Sprintf("📁 %s", fb.currentPath))

	// Clean list without borders or extra styling
	listView := fb.list.View()

	return lipgloss.JoinVertical(lipgloss.Left, header, listView)
}

// ViewWithHelp renders the file browser with help text (for standalone use)
func (fb FileBrowser) ViewWithHelp() string {
	// Simple header with current path
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#569cd6")).
		Bold(true).
		MarginBottom(1).
		Render(fmt.Sprintf("📁 %s", fb.currentPath))

	// Clean list without borders or extra styling
	listView := fb.list.View()

	// Simple help text
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#808080")).
		MarginTop(1).
		Render("Enter: select/navigate • h: toggle hidden • /: filter • Esc: back")

	return lipgloss.JoinVertical(lipgloss.Left, header, listView, helpText)
}

// GetCurrentPath returns the current directory path
func (fb FileBrowser) GetCurrentPath() string {
	return fb.currentPath
}

// GetSelectedPath returns the full path of the selected item
func (fb FileBrowser) GetSelectedPath() string {
	if selectedItem := fb.list.SelectedItem(); selectedItem != nil {
		fileItem := selectedItem.(FileItem)
		if fileItem.name == ".." {
			return filepath.Dir(fb.currentPath)
		}
		return fileItem.path
	}
	return fb.currentPath
}

// GetSelectedItem returns the selected file item
func (fb FileBrowser) GetSelectedItem() *FileItem {
	if selectedItem := fb.list.SelectedItem(); selectedItem != nil {
		fileItem := selectedItem.(FileItem)
		return &fileItem
	}
	return nil
}

// refreshItems scans the current directory and updates the items list
func (fb *FileBrowser) refreshItems() {
	fb.items = []list.Item{}

	// Add "select current directory" option at the top
	fb.items = append(fb.items, FileItem{
		name:  "📍 Select this directory",
		path:  fb.currentPath,
		isDir: true,
	})

	// Add parent directory option if not at root
	if fb.currentPath != "/" && fb.currentPath != filepath.Dir(fb.currentPath) {
		fb.items = append(fb.items, FileItem{
			name:  "..",
			path:  filepath.Dir(fb.currentPath),
			isDir: true,
		})
	}

	// Read directory contents
	entries, err := os.ReadDir(fb.currentPath)
	if err != nil {
		return
	}

	// Separate directories and files
	var dirs, files []FileItem

	for _, entry := range entries {
		name := entry.Name()
		isHidden := strings.HasPrefix(name, ".")

		// Skip hidden files if not showing them
		if isHidden && !fb.showHidden {
			continue
		}

		fullPath := filepath.Join(fb.currentPath, name)
		isDir := entry.IsDir()

		item := FileItem{
			name:     name,
			path:     fullPath,
			isDir:    isDir,
			isHidden: isHidden,
		}

		if isDir {
			dirs = append(dirs, item)
		} else if !fb.directoryOnly {
			files = append(files, item)
		}
	}

	// Sort directories and files separately
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].name) < strings.ToLower(dirs[j].name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].name) < strings.ToLower(files[j].name)
	})

	// Add directories first, then files
	for _, dir := range dirs {
		fb.items = append(fb.items, dir)
	}
	for _, file := range files {
		fb.items = append(fb.items, file)
	}
}

// pathExists checks if a path exists
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
