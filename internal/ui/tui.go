package ui

import (
	"fmt"
	"get-repo/internal/debug"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Forward window size messages to setup wizard when in setup mode
		if m.state == StateSetup {
			m.setupWizard, cmd = m.setupWizard.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			// Update list size for other states
			h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
			width, height := msg.Width-h, msg.Height-v
			if width < 20 {
				width = 20
			}
			if height < 10 {
				height = 10
			}
			m.list.SetSize(width, height)
		}
		
	case tea.KeyMsg:
		// Global key handling
		debug.Log("Key pressed: %s in state %v", msg.String(), m.state)
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		
		// Handle state-specific keys
		switch m.state {
		case StateList:
			return m.handleListKeys(msg)
		case StateClone:
			return m.handleCloneKeys(msg)
		case StateRemoveConfirm:
			return m.handleRemoveConfirmKeys(msg)
		case StateSetup:
			return m.handleSetupKeys(msg)
		case StateUpdateSelection, StateRemoveSelection:
			return m.handleSelectionKeys(msg)
		case StateBatchOperation:
			// No key handling during batch operations
			return m, nil
		}
		
	case cloneFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.statusMsg = "Clone completed successfully!"
			// Refresh the repository list
			return m, m.refreshRepositoryList()
		}
		m.state = StateList
		
	case batchOperationMsg:
		m.operationMutex.Lock()
		m.completedOps++
		m.operationResults = append(m.operationResults, OperationResult{
			RepoName: msg.repoName,
			Success:  msg.success,
			Message:  msg.message,
		})
		
		// Update tree node status
		m.updateNodeStatus(msg.repoName, msg.success, msg.message)
		
		// Update progress
		if m.totalOps > 0 {
			progress := float64(m.completedOps) / float64(m.totalOps)
			cmd = m.progress.SetPercent(progress)
			cmds = append(cmds, cmd)
		}
		
		// Check if all operations are complete
		if m.completedOps >= m.totalOps {
			m.statusMsg = m.generateBatchSummary()
			
			// Clear all selections after batch operation
			items := m.list.Items()
			newItems := make([]list.Item, len(items))
			for i, listItem := range items {
				item := listItem.(Item)
				item.selected = false
				newItems[i] = item
			}
			// Preserve list state when updating items
			currentWidth, currentHeight := m.list.Width(), m.list.Height()
			currentCursor := m.list.Cursor()
			m.list.SetItems(newItems)
			m.list.SetSize(currentWidth, currentHeight)
			m.list.Select(currentCursor)
		}
		m.operationMutex.Unlock()
		
	case refreshListMsg:
		// Just update the title and state without rebuilding the model
		m.state = StateList
		m.list.Title = "Your Repositories"
		return m, nil
		
	case repositoryListMsg:
		// Update the list with new items
		currentWidth, currentHeight := m.list.Width(), m.list.Height()
		currentCursor := m.list.Cursor()
		
		m.list.SetItems(msg.items)
		m.list.SetSize(currentWidth, currentHeight)
		m.list.Title = "Your Repositories"
		
		// Try to maintain cursor position if possible
		if currentCursor < len(msg.items) {
			m.list.Select(currentCursor)
		} else if len(msg.items) > 0 {
			m.list.Select(0)
		}
		
		return m, nil
		
	case error:
		m.err = msg
		return m, nil
	}
	
	// Update sub-components
	switch m.state {
	case StateSetup:
		// Setup wizard handles its own updates
	case StateClone:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	case StateCloning, StateUpdate:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case StateUpdateSelection, StateRemoveSelection:
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	case StateList:
		// Update spinner if operations are running
		if m.totalOps > 0 && m.completedOps < m.totalOps {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		// List navigation is handled in handleListKeys
	case StateBatchOperation:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		newProgress, cmd := m.progress.Update(msg)
		m.progress = newProgress.(progress.Model)
		cmds = append(cmds, cmd)
	}
	
	return m, tea.Batch(cmds...)
}

func (m Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit
	case "c":
		m.state = StateClone
		m.textInput = textinput.New()
		m.textInput.Placeholder = "https://github.com/user/repo"
		m.textInput.Focus()
		m.textInput.Width = 50
		return m, textinput.Blink
	case "u":
		if m.list.SelectedItem() == nil {
			// Go to multi-select mode - preserve current model state
			m.state = StateUpdateSelection
			m.list.Title = "Select repositories to update (Space to toggle, Enter to confirm)"
			
			// Check if we have pre-selected items from the main list
			items := m.list.Items()
			hasSelections := false
			selectedRepos := []string{}
			
			for _, listItem := range items {
				item := listItem.(Item)
				if item.selected && item.isGitRepo {
					hasSelections = true
					selectedRepos = append(selectedRepos, item.node.Path)
				}
			}
			
			// If we have pre-selected items, process them immediately
			if hasSelections {
				// Stay in list state but track operations
				m.totalOps = len(selectedRepos)
				m.completedOps = 0
				m.operationResults = nil // Clear previous results
				m.list.Title = "Your Repositories" // Ensure title is set
				
				// Set pending status for all selected repositories
				for _, repoPath := range selectedRepos {
					m.setNodePending(repoPath)
				}
				
				// Create commands for each repo
				var cmds []tea.Cmd
				for _, repoPath := range selectedRepos {
					cmds = append(cmds, m.updateRepo(repoPath))
				}
				
				cmds = append(cmds, m.spinner.Tick)
				return m, tea.Batch(cmds...)
			}
			
			return m, nil
		}
		// Update single item
		selectedItem := m.list.SelectedItem().(Item)
		// Use the full path from the node for git repos
		repoPath := selectedItem.node.Path
		
		// Set pending status immediately so user sees feedback
		m.setNodePending(repoPath)
		
		// Initialize batch operation tracking for single operation
		m.totalOps = 1
		m.completedOps = 0
		m.operationResults = nil // Clear previous results
		m.list.Title = "Your Repositories" // Ensure title is set
		
		// Stay in list state
		m.statusMsg = fmt.Sprintf("Updating %s...", selectedItem.name)
		return m, tea.Batch(
			m.spinner.Tick, // Start spinner animation
			m.updateRepo(repoPath),
		)
	case "r":
		if m.list.SelectedItem() == nil {
			// Go to multi-select mode - preserve current model state
			m.state = StateRemoveSelection
			m.list.Title = "Select repositories to remove (Space to toggle, Enter to confirm)"
			
			// Check if we have pre-selected items from the main list
			items := m.list.Items()
			hasSelections := false
			selectedRepos := []string{}
			
			for _, listItem := range items {
				item := listItem.(Item)
				if item.selected && item.isGitRepo {
					hasSelections = true
					selectedRepos = append(selectedRepos, item.node.Path)
				}
			}
			
			// If we have pre-selected items, confirm removal
			if hasSelections {
				m.state = StateRemoveConfirm
				m.batchRemoveRepos = selectedRepos
				return m, nil
			}
			
			return m, nil
		}
		// Confirm single removal
		m.state = StateRemoveConfirm
		return m, nil
	case "/":
		// Enable filtering
		m.list.SetFilteringEnabled(true)
		return m, nil
	case " ":
		// Toggle selection for batch operations
		cursor := m.list.Cursor()
		items := m.list.Items()
		
		if cursor < len(items) {
			// Toggle the selected state
			item := items[cursor].(Item)
			item.selected = !item.selected
			
			// Update the item in the list
			newItems := make([]list.Item, len(items))
			copy(newItems, items)
			newItems[cursor] = item
			
			// Preserve list state when updating items
			currentWidth, currentHeight := m.list.Width(), m.list.Height()
			currentCursor := m.list.Cursor()
			m.list.SetItems(newItems)
			m.list.SetSize(currentWidth, currentHeight)
			m.list.Select(currentCursor)
			
			// Update status message with current selection count
			selectedCount := 0
			for _, listItem := range newItems {
				if listItem.(Item).selected {
					selectedCount++
				}
			}
			
			if selectedCount > 0 {
				m.statusMsg = fmt.Sprintf("%d items selected (press 'u' to update, 'r' to remove)", selectedCount)
			} else {
				m.statusMsg = ""
			}
		}
		return m, nil
	case "a":
		// Select all items
		items := m.list.Items()
		newItems := make([]list.Item, len(items))
		for i, listItem := range items {
			item := listItem.(Item)
			item.selected = true
			newItems[i] = item
		}
		// Preserve list state when updating items
		currentWidth, currentHeight := m.list.Width(), m.list.Height()
		currentCursor := m.list.Cursor()
		m.list.SetItems(newItems)
		m.list.SetSize(currentWidth, currentHeight)
		m.list.Select(currentCursor)
		m.statusMsg = fmt.Sprintf("All %d items selected", len(items))
		return m, nil
	case "n":
		// Deselect all items
		items := m.list.Items()
		newItems := make([]list.Item, len(items))
		for i, listItem := range items {
			item := listItem.(Item)
			item.selected = false
			newItems[i] = item
		}
		// Preserve list state when updating items
		currentWidth, currentHeight := m.list.Width(), m.list.Height()
		currentCursor := m.list.Cursor()
		m.list.SetItems(newItems)
		m.list.SetSize(currentWidth, currentHeight)
		m.list.Select(currentCursor)
		m.statusMsg = ""
		return m, nil
	case "right", "l":
		// Expand current item
		return m.handleExpandCollapse(true)
	case "left", "h":
		// Collapse current item
		return m.handleExpandCollapse(false)
	default:
		// Let the list handle navigation keys (up, down, etc.)
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) handleCloneKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		url := m.textInput.Value()
		if url == "" {
			m.err = fmt.Errorf("please enter a URL")
			return m, nil
		}
		m.state = StateCloning
		m.statusMsg = fmt.Sprintf("Cloning %s...", url)
		return m, m.cloneRepo(url)
	case "esc":
		m.state = StateList
		m.err = nil
		return m, nil
	}
	return m, nil
}

func (m Model) handleRemoveConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Check if this is a batch removal
		if len(m.batchRemoveRepos) > 0 {
			// Stay in list state but track operations
			m.state = StateList
			m.totalOps = len(m.batchRemoveRepos)
			m.completedOps = 0
			m.operationResults = nil
			
			// Set pending status for all selected repositories
			for _, repoPath := range m.batchRemoveRepos {
				m.setNodePending(repoPath)
			}
			
			// Create commands for each repo
			var cmds []tea.Cmd
			for _, repoPath := range m.batchRemoveRepos {
				cmds = append(cmds, m.removeRepo(repoPath))
			}
			
			// Clear batch list after starting operation
			m.batchRemoveRepos = nil
			
			cmds = append(cmds, m.spinner.Tick)
			return m, tea.Batch(cmds...)
		}
		
		// Single removal
		selectedItem := m.list.SelectedItem().(Item)
		repoPath := selectedItem.node.Path
		m.state = StateUpdate
		m.statusMsg = fmt.Sprintf("Removing %s...", selectedItem.name)
		return m, m.removeRepo(repoPath)
	default:
		m.state = StateList
		m.batchRemoveRepos = nil // Clear batch list if cancelled
		return m, nil
	}
}

func (m Model) handleSetupKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	// Update setup wizard
	m.setupWizard, cmd = m.setupWizard.Update(msg)
	
	// Check if setup is complete
	if m.setupWizard.step == StepComplete {
		if msg.String() != "" { // Any key press
			// Apply the setup
			if err := m.setupWizard.Apply(); err != nil {
				m.err = fmt.Errorf("setup failed: %w", err)
				return m, nil
			}
			
			// Reinitialize with list state
			return InitialModel(StateList), nil
		}
	}
	
	return m, cmd
}

func (m Model) handleSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case " ":
		// Toggle selection
		index := m.list.Cursor()
		if _, ok := m.selected[index]; ok {
			delete(m.selected, index)
		} else {
			m.selected[index] = struct{}{}
		}
		return m, nil
		
	case "enter":
		// Process selected items
		var selectedRepos []string
		items := m.list.Items()
		
		for idx := range m.selected {
			if idx < len(items) {
				item := items[idx].(Item)
				// Use the full path from the node for git repos
				selectedRepos = append(selectedRepos, item.node.Path)
			}
		}
		
		if len(selectedRepos) == 0 {
			m.state = StateList
			return m, nil
		}
		
		// Stay in list state but track operations
		m.state = StateList
		m.totalOps = len(selectedRepos)
		m.completedOps = 0
		m.operationResults = nil
		
		// Set pending status for all selected repositories
		for _, repoName := range selectedRepos {
			m.setNodePending(repoName)
		}
		
		// Create commands for each repo
		var cmds []tea.Cmd
		for _, repoName := range selectedRepos {
			if m.state == StateUpdateSelection {
				cmds = append(cmds, m.updateRepo(repoName))
			} else {
				cmds = append(cmds, m.removeRepo(repoName))
			}
		}
		
		cmds = append(cmds, m.spinner.Tick)
		return m, tea.Batch(cmds...)
		
	case "esc":
		m.state = StateList
		return m, nil
		
	case "a":
		// Select all
		items := m.list.Items()
		for i := range items {
			m.selected[i] = struct{}{}
		}
		return m, nil
		
	case "n":
		// Select none
		m.selected = make(map[int]struct{})
		return m, nil
	}
	
	return m, nil
}

func (m Model) generateBatchSummary() string {
	successCount := 0
	for _, result := range m.operationResults {
		if result.Success {
			successCount++
		}
	}
	
	failCount := len(m.operationResults) - successCount
	
	if failCount == 0 {
		return fmt.Sprintf("✓ All %d operations completed successfully!", successCount)
	}
	
	return fmt.Sprintf("Completed: %d succeeded, %d failed", successCount, failCount)
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n%s\n\nPress any key to continue...", 
			ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}
	
	var s string
	switch m.state {
	case StateSetup:
		s = m.setupWizard.View()
	case StateClone:
		s = m.renderClone()
	case StateCloning, StateUpdate:
		s = m.renderSpinner()
	case StateRemoveConfirm:
		s = m.renderRemoveConfirm()
	case StateList:
		s = m.renderList()
	case StateUpdateSelection, StateRemoveSelection:
		s = m.renderSelection()
	case StateBatchOperation:
		s = m.renderBatchOperation()
	default:
		s = "Unknown state"
	}
	
	// Add status message if present
	if m.statusMsg != "" && m.state == StateList {
		s += "\n\n" + SuccessStyle.Render(m.statusMsg)
	}
	
	return s
}


func (m Model) renderClone() string {
	return fmt.Sprintf(
		"\n%s\n\nEnter the repository URL to clone.\n\n%s\n\n%s",
		TitleStyle.Render("Clone Repository"),
		m.textInput.View(),
		HelpStyle.Render("Enter to clone • Esc to cancel"),
	)
}

func (m Model) renderSpinner() string {
	return fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.statusMsg)
}

func (m Model) renderRemoveConfirm() string {
	// Check if this is a batch removal
	if len(m.batchRemoveRepos) > 0 {
		return fmt.Sprintf(
			"\n\n   %s\n   This action cannot be undone.\n\n   [y/N]\n\n",
			fmt.Sprintf("Are you sure you want to remove %d repositories?", len(m.batchRemoveRepos)),
		)
	}
	
	// Single removal
	selected := m.list.SelectedItem().(Item).name
	return fmt.Sprintf(
		"\n\n   %s\n   This action cannot be undone.\n\n   [y/N]\n\n",
		fmt.Sprintf("Are you sure you want to remove %s?", TitleStyle.Render(selected)),
	)
}

func (m Model) renderList() string {
	content := lipgloss.NewStyle().Margin(1, 2).Render(m.list.View())
	help := m.getListHelp()
	
	var bottomSection string
	
	// Show operation progress if running
	if m.totalOps > 0 && m.completedOps < m.totalOps {
		progress := fmt.Sprintf("\n\nOperations: %d/%d completed", m.completedOps, m.totalOps)
		bottomSection = PendingStyle.Render(progress)
	}
	
	// Show any recent operation results at the bottom
	if len(m.operationResults) > 0 && m.state == StateList {
		var recentErrors []string
		for _, result := range m.operationResults {
			if !result.Success {
				errorMsg := result.Message
				if len(errorMsg) > 60 {
					errorMsg = errorMsg[:57] + "..."
				}
				recentErrors = append(recentErrors, fmt.Sprintf("  ✗ %s: %s", result.RepoName, errorMsg))
			}
		}
		
		if len(recentErrors) > 0 {
			if bottomSection != "" {
				bottomSection += "\n"
			}
			bottomSection += "\n" + ErrorStyle.Render("Recent Errors:") + "\n" + 
				ErrorStyle.Render(strings.Join(recentErrors, "\n"))
		}
	}
	
	return content + "\n" + help + bottomSection
}

func (m Model) renderSelection() string {
	// Custom render for selection mode
	var lines []string
	items := m.list.Items()
	cursor := m.list.Cursor()
	
	// Add title
	lines = append(lines, TitleStyle.Render(m.list.Title))
	lines = append(lines, "")
	
	// Render items with checkboxes
	start := 0
	if cursor > 10 {
		start = cursor - 10
	}
	end := start + 20
	if end > len(items) {
		end = len(items)
	}
	
	for i := start; i < end; i++ {
		checkbox := "[ ]"
		if _, selected := m.selected[i]; selected {
			checkbox = "[x]"
		}
		
		line := fmt.Sprintf("%s %s", checkbox, items[i].(Item).name)
		
		if i == cursor {
			line = SelectedItemStyle.Render("→ " + line)
		} else {
			line = "  " + line
		}
		
		lines = append(lines, line)
	}
	
	content := lipgloss.NewStyle().Margin(1, 2).Render(strings.Join(lines, "\n"))
	help := m.getSelectionHelp()
	
	return content + "\n\n" + help
}

func (m Model) renderBatchOperation() string {
	// Header section
	header := TitleStyle.Render("Batch Operation")
	
	// Progress section
	progressBar := m.progress.View()
	status := fmt.Sprintf("%s Processing %d/%d repositories...", m.spinner.View(), m.completedOps, m.totalOps)
	
	// Split results into succeeded and failed for better organization
	var succeeded []string
	var failed []string
	var pending []string
	
	// Track which repos have been processed
	processedRepos := make(map[string]bool)
	for _, result := range m.operationResults {
		processedRepos[result.RepoName] = true
		if result.Success {
			succeeded = append(succeeded, fmt.Sprintf("  ✓ %s", result.RepoName))
		} else {
			// Format error message more clearly
			errorMsg := result.Message
			if len(errorMsg) > 50 {
				errorMsg = errorMsg[:47] + "..."
			}
			failed = append(failed, fmt.Sprintf("  ✗ %s", result.RepoName))
			failed = append(failed, fmt.Sprintf("    └─ %s", errorMsg))
		}
	}
	
	// Show pending operations
	items := m.list.Items()
	for _, item := range items {
		if i := item.(Item); i.node != nil && i.node.Status == StatusPending {
			if !processedRepos[i.node.Path] {
				pending = append(pending, fmt.Sprintf("  ⏳ %s", i.node.Name))
			}
		}
	}
	
	// Build content sections
	var sections []string
	sections = append(sections, header)
	sections = append(sections, "")
	sections = append(sections, progressBar)
	sections = append(sections, "")
	sections = append(sections, status)
	sections = append(sections, "")
	
	// Add sections based on what we have
	if len(pending) > 0 {
		sections = append(sections, PendingStyle.Render("Pending:"))
		sections = append(sections, strings.Join(pending, "\n"))
		sections = append(sections, "")
	}
	
	if len(succeeded) > 0 {
		sections = append(sections, SuccessStyle.Render("Succeeded:"))
		sections = append(sections, SuccessStyle.Render(strings.Join(succeeded, "\n")))
		sections = append(sections, "")
	}
	
	if len(failed) > 0 {
		sections = append(sections, ErrorStyle.Render("Failed:"))
		sections = append(sections, ErrorStyle.Render(strings.Join(failed, "\n")))
	}
	
	return "\n" + strings.Join(sections, "\n")
}

func (m Model) getListHelp() string {
	return HelpStyle.Render("↑/↓ navigate • ←/→ collapse/expand • Space select • a all • n none • c clone • u update • r remove • q quit")
}

func (m Model) getSelectionHelp() string {
	selectedCount := len(m.selected)
	status := fmt.Sprintf("%d selected", selectedCount)
	return HelpStyle.Render(fmt.Sprintf("%s • Space toggle • a all • n none • Enter confirm • Esc cancel", status))
}

// handleExpandCollapse handles expanding and collapsing tree nodes
func (m Model) handleExpandCollapse(expand bool) (Model, tea.Cmd) {
	cursor := m.list.Cursor()
	items := m.list.Items()
	
	if cursor >= len(items) {
		return m, nil
	}
	
	item := items[cursor].(Item)
	
	// Only expandable items can be expanded/collapsed
	if !item.isExpandable {
		return m, nil
	}
	
	// Update the node's expanded state
	if expand && !item.node.IsExpanded {
		item.node.IsExpanded = true
	} else if !expand && item.node.IsExpanded {
		item.node.IsExpanded = false
	} else {
		// No change needed
		return m, nil
	}
	
	// Rebuild the tree from the root nodes
	// First, find root nodes by traversing up from current item
	var rootNodes []*TreeNode
	currentNode := item.node
	for currentNode.Parent != nil {
		currentNode = currentNode.Parent
	}
	
	// Collect all root nodes (this is a bit hacky, but works for now)
	// In a better implementation, we'd store root nodes in the model
	allItems := m.list.Items()
	rootMap := make(map[string]*TreeNode)
	
	for _, listItem := range allItems {
		node := listItem.(Item).node
		root := node
		for root.Parent != nil {
			root = root.Parent
		}
		rootMap[root.Name] = root
	}
	
	for _, root := range rootMap {
		rootNodes = append(rootNodes, root)
	}
	
	// Flatten tree and update list
	newItems := flattenTree(rootNodes)
	
	// Preserve list state
	currentWidth, currentHeight := m.list.Width(), m.list.Height()
	m.list.SetItems(newItems)
	m.list.SetSize(currentWidth, currentHeight)
	
	// Try to keep cursor on the same item (by name)
	for i, newItem := range newItems {
		if newItem.(Item).name == item.name && newItem.(Item).level == item.level {
			m.list.Select(i)
			break
		}
	}
	
	return m, nil
}

// updateNodeStatus updates the status of a tree node based on operation result
func (m *Model) updateNodeStatus(repoName string, success bool, message string) {
	items := m.list.Items()
	
	// Find and update the specific node without rebuilding the tree
	for _, listItem := range items {
		item := listItem.(Item)
		
		// Check if this item matches the repository
		if item.isGitRepo && item.node.Path == repoName {
			// Update status directly on the node (this will be reflected in the display)
			if success {
				item.node.Status = StatusSuccess
			} else {
				item.node.Status = StatusFailed
			}
			item.node.StatusMsg = message
			break // Found the item, no need to continue
		}
	}
	
	// Rebuild the display from the tree (preserving expansion states)
	m.refreshTreeDisplay()
}

// setNodePending sets a repository node to pending status
func (m *Model) setNodePending(repoName string) {
	items := m.list.Items()
	
	// Find and update the specific node without rebuilding the tree
	for _, listItem := range items {
		item := listItem.(Item)
		
		// Check if this item matches the repository
		if item.isGitRepo && item.node.Path == repoName {
			item.node.Status = StatusPending
			item.node.StatusMsg = "Operation in progress..."
			break // Found the item, no need to continue
		}
	}
	
	// Rebuild the display from the tree (preserving expansion states)
	m.refreshTreeDisplay()
}

// refreshTreeDisplay rebuilds the flat list from tree nodes while preserving expansion states
func (m *Model) refreshTreeDisplay() {
	items := m.list.Items()
	
	// Collect all root nodes from current items
	rootMap := make(map[string]*TreeNode)
	
	for _, listItem := range items {
		item := listItem.(Item)
		if item.node != nil {
			// Find the root of this node
			root := item.node
			for root.Parent != nil {
				root = root.Parent
			}
			rootMap[root.Name] = root
		}
	}
	
	// Convert map to slice
	var rootNodes []*TreeNode
	for _, root := range rootMap {
		rootNodes = append(rootNodes, root)
	}
	
	// Flatten tree with current expansion states preserved
	newItems := flattenTree(rootNodes)
	
	// Preserve list state
	currentWidth, currentHeight := m.list.Width(), m.list.Height()
	currentCursor := m.list.Cursor()
	
	// Update the list with new items
	m.list.SetItems(newItems)
	m.list.SetSize(currentWidth, currentHeight)
	
	// Try to maintain cursor position if possible
	if currentCursor < len(newItems) {
		m.list.Select(currentCursor)
	} else if len(newItems) > 0 {
		m.list.Select(0)
	}
	
	// Force a refresh to ensure the new status indicators are rendered
	m.list.SetItems(m.list.Items()) // This forces the list to re-render
}

// refreshRepositoryList reloads the repository list from disk
func (m Model) refreshRepositoryList() tea.Cmd {
	return func() tea.Msg {
		// Scan for repositories
		repos, err := m.manager.List()
		if err != nil {
			return error(fmt.Errorf("failed to scan repositories: %w", err))
		}
		
		// Build tree structure from repositories
		tree := buildRepositoryTree(repos)
		
		// Convert tree to flat list for display
		items := flattenTree(tree)
		
		return repositoryListMsg{items: items}
	}
}