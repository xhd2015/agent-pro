package main

import (
	"strings"
	"testing"

	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
)

func TestGrokSessionHelpListsFocus(t *testing.T) {
	if !strings.Contains(grokSessionHelp, groksessions.FocusCommandHelpLine) {
		t.Fatalf("grok session help must include FocusCommandHelpLine %q\n%s", groksessions.FocusCommandHelpLine, grokSessionHelp)
	}
}
