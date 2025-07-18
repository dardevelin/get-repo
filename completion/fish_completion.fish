# Fish completion for get-repo

# Basic commands
complete -c get-repo -f
complete -c get-repo -s h -l help -d "Show help message"
complete -c get-repo -s v -l version -d "Show version information"
complete -c get-repo -s i -l interactive -d "Force interactive TUI mode"
complete -c get-repo -l force -d "Skip confirmation prompts"

# Subcommands
complete -c get-repo -n "__fish_use_subcommand" -a "list" -d "List all repositories"
complete -c get-repo -n "__fish_use_subcommand" -a "update" -d "Update repositories"
complete -c get-repo -n "__fish_use_subcommand" -a "remove" -d "Remove repositories"

# Repository completion for update and remove
complete -c get-repo -n "__fish_seen_subcommand_from update remove" -a "(get-repo list 2>/dev/null)" -d "Repository"

# Force flag for remove command
complete -c get-repo -n "__fish_seen_subcommand_from remove" -l force -d "Skip confirmation prompts"