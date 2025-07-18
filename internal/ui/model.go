package ui

import (
	"fmt"
	"get-repo/config"
	"get-repo/internal/debug"
	"get-repo/internal/repo"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// State represents the current UI state
type State int

const (
	StateList State = iota
	StateSetup
	StateSetupConfigLocation
	StateSetupCodebasesPath
	StateSetupShellIntegration
	StateSetupComplete
	StateClone
	StateCloning
	StateUpdate
	StateRemoveConfirm
	StateUpdateSelection
	StateRemoveSelection
	StateBatchOperation
)

// Model represents the main TUI model
type Model struct {
	state       State
	config      config.Config
	list        list.Model
	textInput   textinput.Model
	spinner     spinner.Model
	progress    progress.Model
	statusMsg   string
	err         error
	selected    map[int]struct{}
	manager     *repo.Manager
	git         *repo.Git
	setupWizard SetupWizard

	// Batch operation tracking
	totalOps         int
	completedOps     int
	operationResults []OperationResult
	operationMutex   sync.Mutex

	// Batch removal tracking
	batchRemoveRepos []string
}

// OperationResult tracks the result of a batch operation
type OperationResult struct {
	RepoName string
	Success  bool
	Message  string
}

// OperationStatus represents the status of an operation on a repository
type OperationStatus int

const (
	StatusNone OperationStatus = iota
	StatusPending
	StatusSuccess
	StatusFailed
)

// TreeNode represents a node in the repository tree
type TreeNode struct {
	Name       string
	Path       string
	IsRepo     bool
	IsExpanded bool
	Level      int
	Children   []*TreeNode
	Parent     *TreeNode
	Status     OperationStatus
	StatusMsg  string
}

// Item represents a list item (flattened tree view)
type Item struct {
	name         string
	selected     bool
	isGitRepo    bool
	isExpandable bool
	isExpanded   bool
	level        int
	node         *TreeNode
	status       OperationStatus
	statusMsg    string
}

func (i Item) Title() string {
	var expandIcon string
	var typeIcon string
	var statusIcon string
	var color string
	var selectionIndicator string

	// Selection indicator - only show for selected items
	if i.selected {
		selectionIndicator = "▸ " // Right-pointing triangle
	} else {
		selectionIndicator = "  "
	}

	// Indentation based on tree level
	indent := strings.Repeat("  ", i.level)

	// Expand/collapse indicator
	if i.isExpandable {
		if i.isExpanded {
			expandIcon = "▼ "
		} else {
			expandIcon = "▶ "
		}
	} else {
		expandIcon = "  "
	}

	// Status indicator - get from node if available
	if i.node != nil {
		i.status = i.node.Status
		i.statusMsg = i.node.StatusMsg
	}

	switch i.status {
	case StatusPending:
		statusIcon = "••• " // Three dots for pending
	case StatusSuccess:
		statusIcon = "✓ " // Simple check mark
	case StatusFailed:
		statusIcon = "✗ " // Simple X mark
	default:
		statusIcon = ""
	}

	// Icon and color based on type and status
	if i.isGitRepo {
		typeIcon = ""
		color = "#4ec9b0" // Git repo color

		// Override color based on status
		switch i.status {
		case StatusSuccess:
			color = "#5fff5f" // Bright green for success
		case StatusFailed:
			color = "#ff5f5f" // Bright red for failure
		case StatusPending:
			color = "#ffff5f" // Yellow for pending
		}
	} else if i.isExpandable {
		typeIcon = ""
		color = "#569cd6" // Directory/organization color
	} else {
		typeIcon = ""
		color = "#569cd6" // Directory color
	}

	// Apply selection styling if selected
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	if i.selected {
		style = style.Bold(true).Background(lipgloss.Color("#264f78")) // Highlight background
	}

	// Build the title with status
	title := fmt.Sprintf("%s%s%s%s%s %s", selectionIndicator, indent, expandIcon, statusIcon, typeIcon, i.name)

	// Don't show error message inline - it's shown at the bottom

	return style.Render(title)
}

func (i Item) Description() string { return "" }
func (i Item) FilterValue() string { return i.name }

// InitialModel creates the initial model
func InitialModel(initialState State) Model {
	defer debug.LogFunction("InitialModel")()
	debug.Log("Creating initial model with state: %v", initialState)

	cfg, err := config.Load()
	if err != nil {
		debug.LogError(err, "loading config in InitialModel")
		return Model{err: err}
	}
	debug.Log("Config loaded: CodebasesPath='%s'", cfg.CodebasesPath)

	// Check if this is a first run
	isFirstRun := config.IsFirstRun()
	debug.Log("IsFirstRun: %v, CodebasesPath empty: %v", isFirstRun, cfg.CodebasesPath == "")

	if isFirstRun || cfg.CodebasesPath == "" {
		debug.Log("Entering setup mode")
		return Model{
			state:       StateSetup,
			config:      cfg,
			setupWizard: NewSetupWizard(),
		}
	}

	// Initialize managers
	debug.Log("Initializing managers with CodebasesPath: %s", cfg.CodebasesPath)
	manager := repo.NewManager(cfg.CodebasesPath)
	git := repo.NewGit(cfg.CodebasesPath)

	// Scan for repositories
	debug.Log("Scanning for repositories...")
	repos, err := manager.List()
	if err != nil {
		debug.LogError(err, "scanning repositories")
		return Model{err: fmt.Errorf("failed to scan repositories: %w", err)}
	}
	debug.Log("Found %d repositories", len(repos))

	// Debug: Check if we found repositories
	if len(repos) == 0 {
		debug.Log("No repositories found in path: %s", cfg.CodebasesPath)
		return Model{err: fmt.Errorf("no repositories found in: %s", cfg.CodebasesPath)}
	}

	// Build tree structure from repositories
	debug.Log("Building repository tree...")
	tree := buildRepositoryTree(repos)
	debug.Log("Built tree with %d root nodes", len(tree))

	// Convert tree to flat list for display
	debug.Log("Flattening tree for display...")
	items := flattenTree(tree)
	debug.Log("Created %d list items", len(items))

	// Create list with proper dimensions
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetHeight(1) // Single line items like file browser
	delegate.SetSpacing(0)

	// Configure delegate styles similar to file browser
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#264f78")).
		Padding(0, 1)
	delegate.Styles.NormalTitle = lipgloss.NewStyle()
	delegate.Styles.DimmedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	delegate.Styles.FilterMatch = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ec9b0"))

	l := list.New(items, delegate, 80, 20) // Start with reasonable size like file browser
	l.Title = "Your Repositories"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)    // Hide status bar like file browser
	l.SetShowTitle(true)         // Show title
	l.SetFilteringEnabled(false) // Disable filtering initially to avoid conflicts
	l.DisableQuitKeybindings()

	// Set title styles to ensure visibility
	l.Styles.Title = TitleStyle

	// Ensure list starts at the top
	if len(items) > 0 {
		l.Select(0)
	}

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ProgressStyle

	// Create progress bar
	p := progress.New(progress.WithDefaultGradient())

	return Model{
		state:    initialState,
		config:   cfg,
		list:     l,
		spinner:  s,
		progress: p,
		selected: make(map[int]struct{}),
		manager:  manager,
		git:      git,
	}
}

func getListTitle(state State) string {
	switch state {
	case StateUpdateSelection:
		return "Select repositories to update (Space to toggle, Enter to confirm)"
	case StateRemoveSelection:
		return "Select repositories to remove (Space to toggle, Enter to confirm)"
	default:
		return "Your Repositories"
	}
}

// Messages
type cloneFinishedMsg struct{ err error }
type updateFinishedMsg struct {
	repoName string
	err      error
}
type removeFinishedMsg struct {
	repoName string
	err      error
}
type batchOperationMsg struct {
	repoName string
	success  bool
	message  string
}
type refreshListMsg struct{}
type repositoryListMsg struct {
	items []list.Item
}

// Commands
func (m Model) cloneRepo(url string) tea.Cmd {
	return func() tea.Msg {
		if err := repo.ValidateURL(url); err != nil {
			return cloneFinishedMsg{err: err}
		}

		clonePath := repo.GetClonePath(url)
		destination := m.manager.GetFullPath(clonePath)

		result := m.git.Clone(url, destination)
		if !result.Success {
			return cloneFinishedMsg{err: result.Error}
		}

		return cloneFinishedMsg{err: nil}
	}
}

func (m Model) updateRepo(repoName string) tea.Cmd {
	return func() tea.Msg {
		debug.Log("updateRepo command starting for: %s", repoName)
		repoPath := m.manager.GetFullPath(repoName)
		debug.Log("Full repo path: %s", repoPath)

		// Mark as pending immediately
		debug.Log("Starting git pull...")
		result := m.git.Pull(repoPath)

		var message string
		if !result.Success {
			// Provide more detailed error information
			if result.Error != nil {
				message = result.Error.Error()
			} else {
				message = "Unknown error occurred"
			}

			// Common git error scenarios
			if strings.Contains(message, "not a git repository") {
				message = "Not a git repository"
			} else if strings.Contains(message, "no such file or directory") {
				message = "Repository path not found"
			} else if strings.Contains(message, "fatal: not a git repository") {
				message = "Invalid git repository"
			} else if strings.Contains(message, "Connection") || strings.Contains(message, "network") {
				message = "Network error - check connection"
			} else if strings.Contains(message, "Permission denied") {
				message = "Permission denied - check credentials"
			} else if strings.Contains(message, "Authentication failed") {
				message = "Authentication failed"
			}

			return batchOperationMsg{
				repoName: repoName,
				success:  false,
				message:  message,
			}
		}

		return batchOperationMsg{
			repoName: repoName,
			success:  true,
			message:  "Updated successfully",
		}
	}
}

func (m Model) removeRepo(repoName string) tea.Cmd {
	return func() tea.Msg {
		repoPath := m.manager.GetFullPath(repoName)

		if err := os.RemoveAll(repoPath); err != nil {
			return batchOperationMsg{
				repoName: repoName,
				success:  false,
				message:  err.Error(),
			}
		}

		return batchOperationMsg{
			repoName: repoName,
			success:  true,
			message:  "Removed successfully",
		}
	}
}

func refreshList() tea.Msg {
	return refreshListMsg{}
}

// buildRepositoryTree creates a hierarchical tree structure from repository list
func buildRepositoryTree(repos []repo.Repository) []*TreeNode {
	var rootNodes []*TreeNode
	nodeMap := make(map[string]*TreeNode)

	// Sort repositories by path for consistent ordering
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})

	for _, r := range repos {
		parts := strings.Split(r.Name, string(filepath.Separator))
		if len(parts) == 0 {
			continue
		}

		var currentNodes []*TreeNode = rootNodes
		var parent *TreeNode = nil
		currentPath := ""

		// Build the path level by level
		for level, part := range parts {
			if currentPath == "" {
				currentPath = part
			} else {
				currentPath = filepath.Join(currentPath, part)
			}

			// Check if node already exists
			var existingNode *TreeNode
			for _, node := range currentNodes {
				if node.Name == part {
					existingNode = node
					break
				}
			}

			if existingNode == nil {
				// Create new node
				isRepo := (level == len(parts)-1) && r.IsGitDir
				node := &TreeNode{
					Name:       part,
					Path:       currentPath,
					IsRepo:     isRepo,
					IsExpanded: level == 0, // VCS providers expanded by default
					Level:      level,
					Parent:     parent,
					Children:   []*TreeNode{},
				}

				if parent == nil {
					rootNodes = append(rootNodes, node)
				} else {
					parent.Children = append(parent.Children, node)
				}

				nodeMap[currentPath] = node
				currentNodes = node.Children
				parent = node
			} else {
				// Use existing node
				currentNodes = existingNode.Children
				parent = existingNode
			}
		}
	}

	// Sort children at each level
	sortTreeNodes(rootNodes)

	return rootNodes
}

// sortTreeNodes recursively sorts tree nodes
func sortTreeNodes(nodes []*TreeNode) {
	sort.Slice(nodes, func(i, j int) bool {
		// VCS providers first, then alphabetical
		if nodes[i].Level == 0 && nodes[j].Level == 0 {
			return nodes[i].Name < nodes[j].Name
		}
		// Repositories last, directories first
		if nodes[i].IsRepo != nodes[j].IsRepo {
			return !nodes[i].IsRepo
		}
		return nodes[i].Name < nodes[j].Name
	})

	for _, node := range nodes {
		sortTreeNodes(node.Children)
	}
}

// flattenTree converts tree structure to flat list for display
func flattenTree(roots []*TreeNode) []list.Item {
	var items []list.Item

	for _, root := range roots {
		flattenNode(root, &items)
	}

	return items
}

// flattenNode recursively flattens a tree node and its children
func flattenNode(node *TreeNode, items *[]list.Item) {
	// Add current node
	item := Item{
		name:         node.Name,
		selected:     false,
		isGitRepo:    node.IsRepo,
		isExpandable: len(node.Children) > 0,
		isExpanded:   node.IsExpanded,
		level:        node.Level,
		node:         node,
		status:       node.Status,
		statusMsg:    node.StatusMsg,
	}
	*items = append(*items, item)

	// Add children if expanded
	if node.IsExpanded {
		for _, child := range node.Children {
			flattenNode(child, items)
		}
	}
}
