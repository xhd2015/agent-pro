package main

import (
	"strings"
	"testing"

	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
)

func TestGrokSessionHelpListsOpen(t *testing.T) {
	if !strings.Contains(grokSessionHelp, groksessions.OpenCommandHelpLine) {
		t.Fatalf("grok session help must include OpenCommandHelpLine %q\n%s", groksessions.OpenCommandHelpLine, grokSessionHelp)
	}
}
