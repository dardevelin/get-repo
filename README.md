# get-repo

A beautiful, feature-rich TUI for managing your git repositories.

`get-repo` provides a fast and interactive way to browse, clone, update, and manage all your local git repositories from a single, elegant terminal interface.

![get-repo-demo](https://user-images.githubusercontent.com/12345/67890.gif) 
*<(This is a placeholder GIF. You can create one using tools like `asciinema` and `agg`)>*

## Features

- **Interactive Repository List:** Browse all your repositories in a filterable, scrollable list.
- **Effortless Cloning:** Press `c`, paste a URL, and clone it directly into your structured directory.
- **Quick Updates:** Update any repository with a single keypress (`u`) to pull the latest changes.
- **Safe Removal:** Remove repositories with a confirmation step (`r`) to prevent accidents.
- **First-Run Setup:** The app guides you through setting up your codebases directory on the first launch.
- **Polished UI:** Built with the power of `charmbracelet/bubbletea` for a modern, responsive terminal experience.

## Installation

Installing `get-repo` is simple with Homebrew.

```sh
# First, tap the formula repository
brew tap dardevelin/get-repo

# Then, install the application
brew install get-repo
```

## Usage

Simply run the application from your terminal:

```sh
get-repo
```

### Keybindings

| Key       | Action                                       |
| :-------- | :------------------------------------------- |
| `↑`/`↓`   | Navigate the repository list                 |
| `c`       | Open the clone view to add a new repository  |
| `u`       | Update the selected repository (`git pull`)    |
| `r`       | Remove the selected repository (with confirm) |
| `q` / `esc` | Quit the application                         |

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
