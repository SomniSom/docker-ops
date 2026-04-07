# bash completion for dq
# Установка: source /path/to/клон-репозитория/completions/bash/dq.bash

_dq_completion() {
	local cur prev cword
	if declare -F _get_comp_words_by_ref >/dev/null; then
		local words
		_get_comp_words_by_ref -n : cur prev words cword
	else
		cur=${COMP_WORDS[COMP_CWORD]}
		prev=${COMP_WORDS[COMP_CWORD - 1]:-}
		cword=$COMP_CWORD
	fi

	local cmds="version validate env config-check build pull up down stop reup ps restart exec status logs logs-tail deploy gen-image-compose completion man help"

	if [[ "$cword" -eq 1 ]]; then
		mapfile -t COMPREPLY < <(compgen -W "$cmds" -- "$cur")
		return
	fi

	local cmd=${COMP_WORDS[1]}
	if [[ "$cmd" == "env" ]]; then
		case "$prev" in
		-o | --output)
			mapfile -t COMPREPLY < <(compgen -f -- "$cur")
			return
			;;
		esac
		mapfile -t COMPREPLY < <(compgen -W "-h --help -o --output -f --force -a --anonymize" -- "$cur")
		return
	fi

	COMPREPLY=()
}

complete -F _dq_completion dq
