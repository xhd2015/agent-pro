package agenttty

import (
	"strings"
	"testing"
)

func TestApplyForkFlagsToArgv(t *testing.T) {
	argv := []string{"grok", "--resume", "parent-sess"}
	// Mirror RunHeadless fork post-process.
	if !hasFlagPair(argv, "--fork-session") {
		argv = append(argv, "--fork-session")
	}
	joined := strings.Join(argv, " ")
	if !strings.Contains(joined, "--resume") || !strings.Contains(joined, "parent-sess") {
		t.Fatal(joined)
	}
	if !strings.Contains(joined, "--fork-session") {
		t.Fatal(joined)
	}
	// Optional forked session id.
	if !hasFlagPair(argv, "--session-id") {
		argv = append(argv, "--session-id", "new-uuid")
	}
	if !strings.Contains(strings.Join(argv, " "), "--session-id") {
		t.Fatal(argv)
	}
}
