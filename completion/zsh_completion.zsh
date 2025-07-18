#compdef get-repo

_get_repo() {
    local context state line
    typeset -A opt_args
    
    _arguments \
        '(-h --help)'{-h,--help}'[Show help message]' \
        '(-v --version)'{-v,--version}'[Show version information]' \
        '(-i --interactive)'{-i,--interactive}'[Force interactive TUI mode]' \
        '--force[Skip confirmation prompts]' \
        '*::command:_get_repo_command'
}

_get_repo_command() {
    local commands repos
    
    commands=(
        'list:List all repositories'
        'update:Update repositories'
        'remove:Remove repositories'
    )
    
    if (( CURRENT == 1 )); then
        # First argument: command or URL
        _alternative \
            'commands:commands:_describe "command" commands' \
            'urls:git urls:_urls'
    elif (( CURRENT == 2 )); then
        case "$words[1]" in
            update|remove)
                # Get repository list
                if (( $+commands[get-repo] )); then
                    repos=(${(f)"$(get-repo list 2>/dev/null)"})
                    _describe 'repositories' repos
                fi
                ;;
        esac
    fi
}

_get_repo "$@"