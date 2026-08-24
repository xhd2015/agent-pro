package main

import (
	"strings"
	"testing"

	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
)

func TestGrokSessionHelpListsSnapshot(t *testing.T) {
	if !strings.Contains(grokSessionHelp, groksessions.SnapshotCommandHelpLine) {
		t.Fatalf("grok session help must include SnapshotCommandHelpLine %q\n%s", groksessions.SnapshotCommandHelpLine, grokSessionHelp)
	}
}
