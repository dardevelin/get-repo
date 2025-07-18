package repo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitOperation represents a git operation result
type GitOperation struct {
	Success bool
	Output  string
	Error   error
}

// Git handles git operations
type Git struct {
	workDir string
}

// NewGit creates a new Git instance
func NewGit(workDir string) *Git {
	return &Git{workDir: workDir}
}

// Clone clones a repository to the specified destination
func (g *Git) Clone(url, destination string) GitOperation {
	// Ensure parent directory exists
	parentDir := filepath.Dir(destination)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return GitOperation{
			Success: false,
			Error:   fmt.Errorf("failed to create directory: %w", err),
		}
	}
	
	cmd := exec.Command("git", "clone", url, destination)
	output, err := g.runCommand(cmd)
	
	return GitOperation{
		Success: err == nil,
		Output:  output,
		Error:   err,
	}
}

// Pull updates a repository
func (g *Git) Pull(repoPath string) GitOperation {
	cmd := exec.Command("git", "-C", repoPath, "pull")
	output, err := g.runCommand(cmd)
	
	return GitOperation{
		Success: err == nil,
		Output:  output,
		Error:   err,
	}
}

// Status gets the status of a repository
func (g *Git) Status(repoPath string) GitOperation {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := g.runCommand(cmd)
	
	return GitOperation{
		Success: err == nil,
		Output:  output,
		Error:   err,
	}
}

// HasUncommittedChanges checks if a repository has uncommitted changes
func (g *Git) HasUncommittedChanges(repoPath string) bool {
	result := g.Status(repoPath)
	return result.Success && strings.TrimSpace(result.Output) != ""
}

// GetRemoteURL gets the remote URL of a repository
func (g *Git) GetRemoteURL(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "config", "--get", "remote.origin.url")
	output, err := g.runCommand(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// IsGitRepository checks if a path is a git repository
func IsGitRepository(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	return err == nil && info.IsDir()
}

// runCommand executes a command and returns combined output
func (g *Git) runCommand(cmd *exec.Cmd) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	output := stdout.String()
	if err != nil {
		if stderr.Len() > 0 {
			output = stderr.String()
		}
		return output, fmt.Errorf("%s: %w", strings.TrimSpace(output), err)
	}
	
	return output, nil
}