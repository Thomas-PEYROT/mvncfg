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
    local cur prev commands
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    commands="list current use init completion help install-completion"

    if [ "$prev" = "use" ]; then
        COMPREPLY=( $(compgen -W "$(mvncfg list)" -- "$cur") )
    elif [ "$prev" = "completion" ]; then
        COMPREPLY=( $(compgen -W "bash zsh" -- "$cur") )
    else
        COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
    fi
}
complete -F _mvncfg mvncfg
`

const zshCompletion = `#compdef mvncfg

_mvncfg() {
    local curcontext=$curcontext state line
    typeset -A opt_args

    _arguments -C \
        '1: :->command' \
        '*::arg:->args'

    case "$state" in
        command)
            _values 'commands' list current use init completion help install-completion
            ;;
        args)
            case "$line[1]" in
                use)
                    _values 'profiles' $(mvncfg list)
                    ;;
                completion)
                    _values 'shells' bash zsh
                    ;;
            esac
            ;;
    esac
}

compdef _mvncfg mvncfg
`
