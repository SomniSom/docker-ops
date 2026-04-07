package sshexec

import "strings"

// ShellQuote returns s safe for POSIX sh single-quoted strings.
func ShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// JoinShell joins already shell-quoted fragments with spaces.
func JoinShell(parts []string) string {
	return strings.Join(parts, " ")
}

// QuoteArgs shell-quotes each argument and joins.
func QuoteArgs(args []string) string {
	q := make([]string, len(args))
	for i, a := range args {
		q[i] = ShellQuote(a)
	}
	return strings.Join(q, " ")
}
