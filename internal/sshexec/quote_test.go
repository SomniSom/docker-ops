package sshexec

import (
	"strings"
	"testing"
)

func TestShellQuote(t *testing.T) {
	if got := ShellQuote(`a'b`); got != `'a'"'"'b'` {
		t.Fatalf("%q", got)
	}
	if got := ShellQuote("plain"); got != `'plain'` {
		t.Fatalf("%q", got)
	}
}

func TestQuoteArgs(t *testing.T) {
	s := QuoteArgs([]string{"compose", "-p", "p1", "-f", "a b.yml"})
	if !strings.Contains(s, `'a b.yml'`) {
		t.Fatalf("%q", s)
	}
}
