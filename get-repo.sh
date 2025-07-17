#!/bin/bash

# This script clones a git repository into a structured directory based on its URL.
# It uses the VCS_CODEBASES environment variable as the root for all repositories.

# Exit immediately if a command exits with a non-zero status.
set -e

# --install flag logic
if [ "$1" = "--install" ]; then
  SCRIPT_PATH=$(realpath "$0")
  SCRIPT_DIR=$(dirname "$SCRIPT_PATH")

  SHELL_CONFIG_FILE=""
  if [ -n "$BASH_VERSION" ]; then
    SHELL_CONFIG_FILE="$HOME/.bashrc"
  elif [ -n "$ZSH_VERSION" ]; then
    SHELL_CONFIG_FILE="$HOME/.zshrc"
  else
    echo "Unsupported shell. Only bash and zsh are supported for --install."
    exit 1
  fi

  # Add get-repo to PATH
  if ! grep -q "export PATH=\"$SCRIPT_DIR:\$PATH\"" "$SHELL_CONFIG_FILE"; then
    echo "Adding get-repo to your PATH in $SHELL_CONFIG_FILE"
    echo '' >> "$SHELL_CONFIG_FILE"
    echo '# Added by get-repo --install' >> "$SHELL_CONFIG_FILE"
    echo "export PATH=\"$SCRIPT_DIR:\$PATH\"" >> "$SHELL_CONFIG_FILE"
  else
    echo "get-repo is already in your PATH."
  fi

  # Handle VCS_CODEBASES
  VCS_CODEBASES_PATH=$2
  DEFAULT_VCS_PATH="$HOME/dev/vcs-codebases"

  if [ -z "$VCS_CODEBASES_PATH" ]; then
    read -e -p "Enter the path for your codebases directory [default: $DEFAULT_VCS_PATH]: " VCS_CODEBASES_PATH
  fi
  
  # If user just presses enter, the variable might be empty, so re-assign default
  VCS_CODEBASES_PATH=${VCS_CODEBASES_PATH:-$DEFAULT_VCS_PATH}

  # Check if the directory exists, if not, offer to create it.
  if [ ! -d "$VCS_CODEBASES_PATH" ]; then
    read -p "Directory '$VCS_CODEBASES_PATH' does not exist. Create it? [y/N] " confirm
    if [[ "${confirm,,}" == "y" ]]; then
      echo "Creating directory: $VCS_CODEBASES_PATH"
      mkdir -p "$VCS_CODEBASES_PATH"
    else
      echo "Directory not created. Please create it manually and run the installer again."
      exit 1
    fi
  fi

  # Set VCS_CODEBASES in shell config
  if ! grep -q "export VCS_CODEBASES=" "$SHELL_CONFIG_FILE"; then
    echo "Setting VCS_CODEBASES in $SHELL_CONFIG_FILE"
    echo '' >> "$SHELL_CONFIG_FILE"
    echo '# Set by get-repo --install' >> "$SHELL_CONFIG_FILE"
    echo "export VCS_CODEBASES=\"$VCS_CODEBASES_PATH\"" >> "$SHELL_CONFIG_FILE"
  else
    echo "VCS_CODEBASES is already set in $SHELL_CONFIG_FILE."
  fi
  
  echo "Installation complete. Please restart your shell or run 'source $SHELL_CONFIG_FILE'."
  exit 0
fi

# Check if the repository URL is provided as an argument.
if [ -z "$1" ]; then
  echo "Usage: $0 <git-repo-url> or $0 --install"
  exit 1
fi


REPO_URL=$1

# Check if the VCS_CODEBASES environment variable is set.
if [ -z "$VCS_CODEBASES" ]; then
  echo "Error: The VCS_CODEBASES environment variable is not set."
  echo "Please set it to the absolute path of your codebases directory."
  exit 1
fi

# Ensure the VCS_CODEBASES directory exists.
if [ ! -d "$VCS_CODEBASES" ]; then
  echo "Error: The directory specified by VCS_CODEBASES does not exist: $VCS_CODEBASES"
  exit 1
fi

# Parse the URL to create the target directory path.
# 1. Remove protocol (https://, http://, git@).
# 2. Replace the first colon (for SSH URLs like git@host:user/repo) with a slash.
# 3. Remove the trailing .git extension if it exists.
CLONE_PATH=$(echo "$REPO_URL" | sed -e 's#^https://##' -e 's#^http://##' -e 's#^git@##' | sed 's#:#/#' | sed 's#\.git$##')

# The final destination for the repository.
REPO_DEST="$VCS_CODEBASES/$CLONE_PATH"
# The directory where the repository will be cloned.
TARGET_DIR=$(dirname "$REPO_DEST")

# Check if the repository has already been cloned.
if [ -d "$REPO_DEST" ]; then
  echo "Repository already exists at: $REPO_DEST"
  exit 0
fi

# Create the target directory structure.
echo "Creating directory: $TARGET_DIR"
mkdir -p "$TARGET_DIR"

# Clone the repository into the target destination.
echo "Cloning $REPO_URL into $REPO_DEST"
git clone "$REPO_URL" "$REPO_DEST"

echo "Repository cloned successfully."
