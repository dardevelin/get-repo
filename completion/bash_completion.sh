#!/bin/bash

_get_repo_completion() {
    local cur prev opts repo_list
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Basic commands and options
    opts="list update remove --help --version --interactive --force"
    
    case "${prev}" in
        update|remove)
            # Get repository list for update/remove commands
            if command -v get-repo >/dev/null 2>&1; then
                repo_list=$(get-repo list 2>/dev/null)
                COMPREPLY=($(compgen -W "${repo_list}" -- ${cur}))
                return 0
            fi
            ;;
        get-repo)
            # Complete with commands, URLs, or repository names
            COMPREPLY=($(compgen -W "${opts}" -- ${cur}))
            
            # Add URL completion hints
            if [[ ${cur} == http* ]] || [[ ${cur} == git@* ]]; then
                # Don't complete URLs, let user type them
                return 0
            fi
            
            return 0
            ;;
    esac
    
    # Default completion
    COMPREPLY=($(compgen -W "${opts}" -- ${cur}))
}

# Register the completion function
complete -F _get_repo_completion get-repo