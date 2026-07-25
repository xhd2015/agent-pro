package procresolve

import (
	"path/filepath"
	"strings"
)

// classifyCmd maps a full command line (path + argv) to a role other than
// "input". "input" is applied only to the requested root pid in the tree.
//
// Grok update utilities are not session runners and map to "other".
func classifyCmd(cmd string) string {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return "other"
	}
	base := filepath.Base(fields[0])

	switch {
	case base == "agent-run":
		for _, f := range fields[1:] {
			if f == "serve" {
				return "agent-run-serve"
			}
			if strings.HasPrefix(f, "-") {
				break
			}
		}
		return "agent-run"
	case base == "grok":
		if isGrokUpdate(fields) {
			return "other"
		}
		return "grok"
	case base == "codex":
		return "codex"
	default:
		return "other"
	}
}

// isGrokUpdate reports whether argv looks like `grok update …`.
func isGrokUpdate(fields []string) bool {
	if len(fields) < 2 {
		return false
	}
	if filepath.Base(fields[0]) != "grok" {
		return false
	}
	for _, f := range fields[1:] {
		if strings.HasPrefix(f, "-") {
			continue
		}
		return f == "update"
	}
	return false
}

// runnerKind returns "grok" or "codex" if cmd is a session runner candidate,
// else "".
func runnerKind(cmd string) string {
	role := classifyCmd(cmd)
	switch role {
	case "grok", "codex":
		return role
	default:
		return ""
	}
}
