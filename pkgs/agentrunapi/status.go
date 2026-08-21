package agentrunapi

import "strings"

// ParseTTYStatus extracts screen status and sendable token from human
// `agent-run tty status` stdout (parity with agentrunbridge.ParseTTYStatus).
//
// Lines:
//   - "screen status: <value>" → screen (trimmed value after colon)
//   - "sendable: <token> ..." → first whitespace-separated token of the value
func ParseTTYStatus(stdout string) (screen, sendable string) {
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "screen status:"):
			if i := strings.Index(line, ":"); i >= 0 {
				screen = strings.TrimSpace(line[i+1:])
			}
		case strings.HasPrefix(lower, "sendable:"):
			if i := strings.Index(line, ":"); i >= 0 {
				val := strings.TrimSpace(line[i+1:])
				fields := strings.Fields(val)
				if len(fields) > 0 {
					sendable = fields[0]
				}
			}
		}
	}
	return screen, sendable
}

// ParseRunnerSessionIDFromStatus extracts the provider runner session id from
// `agent-run tty status` human stdout ("runner session id: <id>").
// Returns "" when unbound / missing.
func ParseRunnerSessionIDFromStatus(stdout string) string {
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if !strings.HasPrefix(lower, "runner session id:") {
			continue
		}
		i := strings.Index(line, ":")
		if i < 0 {
			return ""
		}
		val := strings.TrimSpace(line[i+1:])
		if val == "" || strings.EqualFold(val, "(unbound)") || strings.EqualFold(val, "unbound") {
			return ""
		}
		return val
	}
	return ""
}

// IsSessionReadyFromStatus reports sendable yes at an idle/banner prompt with a
// bound runner_session_id when the status line is present.
//
// Grok TTY reports screen=idle at the sendable composer; older codex paths may
// still say banner. When "runner session id:" is absent (older binaries), bind
// is not required for backward compatibility.
func IsSessionReadyFromStatus(stdout string) bool {
	screen, sendable := ParseTTYStatus(stdout)
	if !strings.EqualFold(sendable, "yes") {
		return false
	}
	screenOK := strings.EqualFold(screen, "banner") || strings.EqualFold(screen, "idle")
	if !screenOK {
		return false
	}
	if !statusHasRunnerSessionIDLine(stdout) {
		return true
	}
	return ParseRunnerSessionIDFromStatus(stdout) != ""
}

func statusHasRunnerSessionIDLine(stdout string) bool {
	for _, line := range strings.Split(stdout, "\n") {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "runner session id:") {
			return true
		}
	}
	return false
}
