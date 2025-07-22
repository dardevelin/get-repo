# Changelog

All notable changes to this project will be documented in this file.

## [1.0.3] - 2025-07-22

### Added
- Short notation support with fuzzy matching for popular git hosting services:
  - Common abbreviations: `gh:`, `gl:`, `bb:` for GitHub, GitLab, and Bitbucket
  - Full names: `github:`, `gitlab:`, `bitbucket:`
  - Partial matches: `git:` (GitHub), `gitl:` (GitLab), `bit:` (Bitbucket)
  - Fuzzy matching automatically expands any prefix that uniquely identifies a provider
- `--cd` flag that outputs the repository path after clone/update operations
  - Use with command substitution: `cd $(get-repo gh:user/repo --cd)`
- Short notation support in bulk clone files (`-f` option)

### Changed
- Updated help text and man pages to document new features
- Enhanced URL validation to support short notation

## [1.0.2] - Previous release
- Various bug fixes and improvements

## [1.0.1] - Previous release
- Fixed zsh completion syntax errors

## [1.0.0] - Initial release
- Core functionality for managing git repositories
- Interactive TUI for browsing repositories
- Bulk clone support
- Shell completions