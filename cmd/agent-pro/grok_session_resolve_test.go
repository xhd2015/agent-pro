package main

import (
	"strings"
	"testing"

	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
)

func TestGrokSessionHelpListsResolve(t *testing.T) {
	if !strings.Contains(grokSessionHelp, groksessions.ResolveCommandHelpLine) {
		t.Fatalf("grok session help must include ResolveCommandHelpLine %q\n%s", groksessions.ResolveCommandHelpLine, grokSessionHelp)
	}
}

func TestGrokSessionHelpListsFork(t *testing.T) {
	if !strings.Contains(grokSessionHelp, groksessions.ForkCommandHelpLine) {
		t.Fatalf("grok session help must include ForkCommandHelpLine %q\n%s", groksessions.ForkCommandHelpLine, grokSessionHelp)
	}
}
