// Package completion provides shell completion scripts for mvncfg.
package completion

// Bash returns the bash completion script.
func Bash() string {
	return bashCompletion
}

// Zsh returns the zsh completion script.
func Zsh() string {
	return zshCompletion
}

const bashCompletion = `_mvncfg() {
    local cur prev commands shells
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    commands="init list current use create install-completion completion help"
    shells="bash zsh"

    if [ "$prev" = "use" ]; then
        COMPREPLY=( $(compgen -W "$(mvncfg list)" -- "$cur") )
    elif [ "$prev" = "completion" ]; then
        COMPREPLY=( $(compgen -W "$shells" -- "$cur") )
    elif [ "$prev" = "help" ]; then
        COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
    else
        COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
    fi
}
complete -F _mvncfg mvncfg
`

const zshCompletion = `#compdef mvncfg

local -a cmd_list shell_list
cmd_list=(
    'init:initialize the profiles directory and default profile'
    'list:list available profiles'
    'current:show the active profile'
    'use:activate a profile'
    'create:create a new profile from a default template'
    'install-completion:install shell completion'
    'completion:print the raw completion script'
    'help:show help for a command'
)
shell_list=(bash zsh)

_mvncfg() {
    local curcontext=$curcontext state line
    typeset -A opt_args

    _arguments -C \
        '1: :->command' \
        '*::arg:->args'

    case "$state" in
        command)
            _describe -t commands 'mvncfg commands' cmd_list
            ;;
        args)
            case "$line[1]" in
                use)
                    _values 'profiles' $(mvncfg list)
                    ;;
                completion)
                    _describe -t shells 'shell' shell_list
                    ;;
                help)
                    _describe -t commands 'command' cmd_list
                    ;;
            esac
            ;;
    esac
}

compdef _mvncfg mvncfg
`
