package agentrunbridge

import "strings"

// ParseTTYStatus extracts screen status and sendable token from
// `agent-run tty status` human stdout.
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

// IsSessionReady reports whether tty status stdout is idle/banner + sendable yes
// with a bound runner_session_id when that line is present (parity with
// agentrunapi.IsSessionReadyFromStatus).
func IsSessionReady(stdout string) bool {
	screen, sendable := ParseTTYStatus(stdout)
	if !strings.EqualFold(sendable, "yes") {
		return false
	}
	screenOK := strings.EqualFold(screen, "banner") || strings.EqualFold(screen, "idle")
	if !screenOK {
		return false
	}
	hasBindLine := false
	for _, line := range strings.Split(stdout, "\n") {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "runner session id:") {
			hasBindLine = true
			break
		}
	}
	if !hasBindLine {
		return true
	}
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)
		if !strings.HasPrefix(lower, "runner session id:") {
			continue
		}
		i := strings.Index(line, ":")
		if i < 0 {
			return false
		}
		val := strings.TrimSpace(line[i+1:])
		if val == "" || strings.EqualFold(val, "(unbound)") || strings.EqualFold(val, "unbound") {
			return false
		}
		return true
	}
	return false
}
