# Contributing to get-repo

Thank you for your interest in contributing to get-repo! This guide will help you get started.

## Development Setup

### Prerequisites

1. **Go 1.20 or higher**
   ```bash
   brew install go
   ```

2. **Git**
   ```bash
   brew install git
   ```

3. **go-md2man** (for man page generation)
   ```bash
   brew install go-md2man
   ```

4. **golangci-lint** (optional, for linting)
   ```bash
   brew install golangci-lint
   ```

### Getting Started

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/get-repo.git
   cd get-repo
   ```

3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/dardevelin/get-repo.git
   ```

4. Install dependencies:
   ```bash
   make deps
   ```

5. Build the project:
   ```bash
   make build
   ```

## Development Workflow

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build with debug support
go build -tags debug -o get-repo-debug ./cmd/get-repo
```

### Testing

```bash
# Run tests
make test

# Run linter
make lint
```

### Man Page

```bash
# Generate man page
make man

# View man page
make view-man
```

### Debugging

Enable debug logging by building with debug tags:
```bash
go build -tags debug -o get-repo-debug ./cmd/get-repo
./get-repo-debug
```

Debug logs will be written to `debug.log`.

## Code Structure

```
get-repo/
├── cmd/get-repo/         # Main entry point
├── internal/            # Internal packages (not exported)
│   ├── cli/            # CLI parsing and command execution
│   ├── debug/          # Debug logging utilities
│   ├── repo/           # Repository management logic
│   └── ui/             # Terminal UI components
├── pkg/                # Public packages
│   └── version/        # Version information
├── config/             # Configuration management
├── docs/               # Documentation and man pages
└── completion/         # Shell completion scripts
```

## Making Changes

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the code style

3. Test your changes:
   ```bash
   make test
   make build
   ./get-repo  # Manual testing
   ```

4. Commit your changes:
   ```bash
   git add .
   git commit -m "Brief description of change"
   ```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions focused and small
- Add comments for exported functions
- Avoid adding unnecessary comments

## Pull Request Process

1. Update your fork:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

3. Create a pull request on GitHub

4. Ensure all checks pass

5. Wait for review

## Release Process

Releases are managed by maintainers:

1. Update version in `pkg/version/version.go`
2. Update CHANGELOG (if exists)
3. Commit with message "Release version X.Y.Z"
4. Tag the release: `git tag vX.Y.Z`
5. Push tag: `git push origin vX.Y.Z`

## Need Help?

- Open an issue on GitHub
- Check existing issues and pull requests
- Read the codebase - it's not too large!

## License

By contributing, you agree that your contributions will be licensed under the MIT License.