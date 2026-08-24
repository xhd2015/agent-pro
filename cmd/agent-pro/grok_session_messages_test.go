package main

import (
	"strings"
	"testing"

	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
)

func TestGrokSessionHelpListsMessages(t *testing.T) {
	if !strings.Contains(grokSessionHelp, groksessions.MessagesCommandHelpLine) {
		t.Fatalf("grok session help must include MessagesCommandHelpLine %q\n%s", groksessions.MessagesCommandHelpLine, grokSessionHelp)
	}
}
