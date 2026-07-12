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

// IsSessionReady reports whether tty status stdout is banner + sendable yes.
func IsSessionReady(stdout string) bool {
	screen, sendable := ParseTTYStatus(stdout)
	return strings.EqualFold(screen, "banner") && strings.EqualFold(sendable, "yes")
}
