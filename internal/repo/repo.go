package repo

import (
	"fmt"
	"get-repo/internal/debug"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Repository represents a git repository
type Repository struct {
	Name     string
	Path     string
	URL      string
	IsGitDir bool
}

// Manager handles repository operations
type Manager struct {
	basePath string
}

// NewManager creates a new repository manager
func NewManager(basePath string) *Manager {
	return &Manager{basePath: basePath}
}

// List returns all directories and repositories found under the base path
func (m *Manager) List() ([]Repository, error) {
	defer debug.LogFunction("Manager.List")()
	debug.Log("Scanning base path: %s", m.basePath)

	var repos []Repository
	visited := make(map[string]bool) // Track visited paths to avoid duplicates

	// Check if base path exists
	if _, err := os.Stat(m.basePath); os.IsNotExist(err) {
		debug.Log("Base path does not exist: %s", m.basePath)
		return nil, fmt.Errorf("base path does not exist: %s", m.basePath)
	}

	// First pass: Find all git repositories
	debug.Log("First pass: scanning for git repositories...")
	err := filepath.WalkDir(m.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			debug.LogError(err, fmt.Sprintf("walking path %s", path))
			return err
		}

		if d.IsDir() && d.Name() == ".git" {
			repoPath := filepath.Dir(path)
			relPath, err := filepath.Rel(m.basePath, repoPath)
			if err != nil {
				debug.LogError(err, fmt.Sprintf("getting relative path for %s", repoPath))
				return err
			}

			debug.Log("Found git repository: %s", relPath)
			repos = append(repos, Repository{
				Name:     relPath,
				Path:     repoPath,
				IsGitDir: true,
			})
			visited[relPath] = true

			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		debug.LogError(err, "first pass walkdir")
		return nil, err
	}
	debug.Log("First pass complete, found %d git repositories", len(repos))

	// Second pass: Find all directories (including organizational folders)
	debug.Log("Second pass: scanning for organizational directories...")
	err = filepath.WalkDir(m.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			debug.LogError(err, fmt.Sprintf("walking path %s in second pass", path))
			return err
		}

		// Skip the base path itself
		if path == m.basePath {
			return nil
		}

		if d.IsDir() && d.Name() != ".git" {
			relPath, err := filepath.Rel(m.basePath, path)
			if err != nil {
				debug.LogError(err, fmt.Sprintf("getting relative path for %s in second pass", path))
				return err
			}

			// Skip if we already found this as a git repository
			if visited[relPath] {
				debug.Log("Skipping already visited path: %s", relPath)
				return filepath.SkipDir
			}

			// Skip hidden directories and nested paths of git repos
			if strings.HasPrefix(d.Name(), ".") {
				debug.Log("Skipping hidden directory: %s", relPath)
				return filepath.SkipDir
			}

			// Only show top-level directories and direct subdirectories
			pathParts := strings.Split(relPath, string(filepath.Separator))
			if len(pathParts) <= 2 { // e.g., "gitlab.com" or "github.com/user"
				debug.Log("Found organizational directory: %s", relPath)
				repos = append(repos, Repository{
					Name:     relPath,
					Path:     path,
					IsGitDir: false,
				})
				visited[relPath] = true
			}
		}

		return nil
	})

	if err != nil {
		debug.LogError(err, "second pass walkdir")
	}
	debug.Log("Repository scan complete, total found: %d", len(repos))
	return repos, err
}

// ExpandShortNotation expands short notation like gh:user/repo to full URLs
func ExpandShortNotation(input string) string {
	// Check if input contains colon for short notation
	colonIndex := strings.Index(input, ":")
	if colonIndex == -1 || colonIndex == 0 {
		return input
	}

	// Extract prefix and path
	prefix := strings.ToLower(input[:colonIndex])
	path := input[colonIndex+1:]

	// Define known providers
	providers := map[string]string{
		"github":    "github.com",
		"gitlab":    "gitlab.com",
		"bitbucket": "bitbucket.org",
	}

	// Try exact match first
	for name, domain := range providers {
		if prefix == name {
			return fmt.Sprintf("https://%s/%s", domain, path)
		}
	}

	// Fuzzy match - find the provider that starts with the prefix
	var matches []string
	var matchedDomains []string
	
	for name, domain := range providers {
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, name)
			matchedDomains = append(matchedDomains, domain)
		}
	}

	// If exactly one match, use it
	if len(matches) == 1 {
		return fmt.Sprintf("https://%s/%s", matchedDomains[0], path)
	}

	// If no matches or multiple matches, try common abbreviations
	commonAbbreviations := map[string]string{
		"gh": "github.com",
		"gl": "gitlab.com",
		"bb": "bitbucket.org",
		"git": "github.com", // Default "git" to GitHub as it's most common
	}

	if domain, ok := commonAbbreviations[prefix]; ok {
		return fmt.Sprintf("https://%s/%s", domain, path)
	}

	// Return input unchanged if no clear match
	return input
}

// ValidateURL validates and normalizes a git repository URL
func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("empty URL provided")
	}

	// Expand short notation first
	expandedURL := ExpandShortNotation(rawURL)

	// Handle SSH URLs
	if strings.HasPrefix(expandedURL, "git@") {
		return nil
	}

	// Handle HTTP(S) URLs
	if strings.HasPrefix(expandedURL, "http://") || strings.HasPrefix(expandedURL, "https://") {
		_, err := url.Parse(expandedURL)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
		return nil
	}

	// Handle SCP-style URLs
	if strings.Contains(expandedURL, ":") && !strings.Contains(expandedURL, "://") {
		return nil
	}

	return fmt.Errorf("unsupported URL format: %s", expandedURL)
}

// GetClonePath derives the local filesystem path from a git URL
func GetClonePath(gitURL string) string {
	// Expand short notation first
	path := ExpandShortNotation(gitURL)

	// Remove protocol prefixes
	path = strings.TrimPrefix(path, "https://")
	path = strings.TrimPrefix(path, "http://")
	path = strings.TrimPrefix(path, "git@")

	// Handle SSH URLs (git@host:user/repo)
	if strings.Contains(path, ":") && !strings.Contains(path, "://") {
		path = strings.Replace(path, ":", "/", 1)
	}

	// Remove .git suffix
	path = strings.TrimSuffix(path, ".git")

	return path
}

// PathExists checks if a repository path already exists
func (m *Manager) PathExists(repoName string) bool {
	fullPath := filepath.Join(m.basePath, repoName)
	_, err := os.Stat(fullPath)
	return err == nil
}

// GetFullPath returns the full filesystem path for a repository
func (m *Manager) GetFullPath(repoName string) string {
	return filepath.Join(m.basePath, repoName)
}
