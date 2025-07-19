#!/usr/bin/env bash
# Development setup script for get-repo

set -e

echo "Setting up get-repo development environment..."

# Check for required tools
check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo "❌ $1 is not installed"
        return 1
    else
        echo "✓ $1 is installed"
        return 0
    fi
}

echo -e "\nChecking prerequisites..."
missing_deps=0

if ! check_command "go"; then
    echo "  Install with: brew install go"
    missing_deps=1
fi

if ! check_command "git"; then
    echo "  Install with: brew install git"
    missing_deps=1
fi

if ! check_command "go-md2man"; then
    echo "  Install with: brew install go-md2man"
    missing_deps=1
fi

if ! check_command "golangci-lint"; then
    echo "  Install with: brew install golangci-lint (optional)"
fi

if [ $missing_deps -eq 1 ]; then
    echo -e "\n❌ Please install missing dependencies first"
    exit 1
fi

echo -e "\n✓ All required dependencies are installed"

# Install Go dependencies
echo -e "\nInstalling Go dependencies..."
make deps

# Build the project
echo -e "\nBuilding get-repo..."
make build

# Generate man page
echo -e "\nGenerating man page..."
make man

echo -e "\n✅ Development environment setup complete!"
echo -e "\nYou can now:"
echo "  - Run 'make build' to build the project"
echo "  - Run 'make test' to run tests"
echo "  - Run 'make lint' to check code style"
echo "  - Run './get-repo' to test the application"