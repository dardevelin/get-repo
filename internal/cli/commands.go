package cli

import (
	"bufio"
	"fmt"
	"get-repo/config"
	"get-repo/internal/repo"
	"os"
	"strings"
	"sync"
)

// Runner handles non-interactive command execution
type Runner struct {
	config  config.Config
	manager *repo.Manager
	git     *repo.Git
}

// NewRunner creates a new command runner
func NewRunner(cfg config.Config) *Runner {
	return &Runner{
		config:  cfg,
		manager: repo.NewManager(cfg.CodebasesPath),
		git:     repo.NewGit(cfg.CodebasesPath),
	}
}

// List lists all repositories
func (r *Runner) List() error {
	repos, err := r.manager.List()
	if err != nil {
		return fmt.Errorf("error scanning repositories: %w", err)
	}
	
	if len(repos) == 0 {
		fmt.Println("No repositories found.")
		return nil
	}
	
	for _, repo := range repos {
		fmt.Println(repo.Name)
	}
	
	return nil
}

// Clone clones a repository
func (r *Runner) Clone(url string) error {
	// Validate URL
	if err := repo.ValidateURL(url); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	
	// Get destination path
	clonePath := repo.GetClonePath(url)
	destination := r.manager.GetFullPath(clonePath)
	
	// Check if already exists
	if r.manager.PathExists(clonePath) {
		return fmt.Errorf("repository already exists at %s", clonePath)
	}
	
	fmt.Printf("Cloning %s into %s...\n", url, clonePath)
	
	// Perform clone
	result := r.git.Clone(url, destination)
	if !result.Success {
		return fmt.Errorf("clone failed: %w", result.Error)
	}
	
	fmt.Println("Clone completed successfully.")
	return nil
}

// Update updates one or more repositories
func (r *Runner) Update(repoNames []string) error {
	if len(repoNames) == 0 {
		return fmt.Errorf("no repositories specified")
	}
	
	if len(repoNames) == 1 {
		return r.updateSingle(repoNames[0])
	}
	
	return r.updateMultiple(repoNames)
}

// updateSingle updates a single repository
func (r *Runner) updateSingle(repoName string) error {
	repoPath := r.manager.GetFullPath(repoName)
	
	if !repo.IsGitRepository(repoPath) {
		return fmt.Errorf("%s is not a git repository", repoName)
	}
	
	fmt.Printf("Updating %s...\n", repoName)
	
	result := r.git.Pull(repoPath)
	if !result.Success {
		return fmt.Errorf("update failed: %w", result.Error)
	}
	
	fmt.Println("Update completed successfully.")
	if result.Output != "" {
		fmt.Println(strings.TrimSpace(result.Output))
	}
	
	return nil
}

// updateMultiple updates multiple repositories in parallel
func (r *Runner) updateMultiple(repoNames []string) error {
	var wg sync.WaitGroup
	results := make(chan updateResult, len(repoNames))
	
	for _, repoName := range repoNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			
			repoPath := r.manager.GetFullPath(name)
			if !repo.IsGitRepository(repoPath) {
				results <- updateResult{
					repoName: name,
					success:  false,
					err:      fmt.Errorf("not a git repository"),
				}
				return
			}
			
			result := r.git.Pull(repoPath)
			results <- updateResult{
				repoName: name,
				success:  result.Success,
				output:   result.Output,
				err:      result.Error,
			}
		}(repoName)
	}
	
	// Wait for all updates to complete
	wg.Wait()
	close(results)
	
	// Print results
	successCount := 0
	failCount := 0
	
	fmt.Println("\nUpdate Results:")
	fmt.Println(strings.Repeat("-", 50))
	
	for result := range results {
		if result.success {
			successCount++
			fmt.Printf("✓ %s: Updated successfully\n", result.repoName)
		} else {
			failCount++
			fmt.Printf("✗ %s: Failed - %v\n", result.repoName, result.err)
		}
	}
	
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Summary: %d succeeded, %d failed\n", successCount, failCount)
	
	if failCount > 0 {
		return fmt.Errorf("%d updates failed", failCount)
	}
	
	return nil
}

// Remove removes one or more repositories
func (r *Runner) Remove(repoNames []string, force bool) error {
	if len(repoNames) == 0 {
		return fmt.Errorf("no repositories specified")
	}
	
	// Verify all repos exist first
	for _, repoName := range repoNames {
		if !r.manager.PathExists(repoName) {
			return fmt.Errorf("repository %s not found", repoName)
		}
	}
	
	// Confirm removal if not forced
	if !force {
		fmt.Printf("Are you sure you want to remove the following repositories?\n")
		for _, name := range repoNames {
			fmt.Printf("  - %s\n", name)
		}
		fmt.Print("\nThis action cannot be undone. Continue? [y/N] ")
		
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(input)) != "y" {
			fmt.Println("Remove cancelled.")
			return nil
		}
	}
	
	// Remove repositories
	for _, repoName := range repoNames {
		repoPath := r.manager.GetFullPath(repoName)
		fmt.Printf("Removing %s...\n", repoName)
		
		if err := os.RemoveAll(repoPath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", repoName, err)
		}
	}
	
	fmt.Printf("Successfully removed %d repositories.\n", len(repoNames))
	return nil
}

type updateResult struct {
	repoName string
	success  bool
	output   string
	err      error
}