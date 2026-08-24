package knowledgesink

import (
	"fmt"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

// EffectiveAgentRunner maps shared/prefs runner IDs to the launch family for mode.
//
// Headless uses CLI runners (codex/grok); --open uses TTY (*-tty).
// Returns mappedFrom when an alias was applied (for notices).
func EffectiveAgentRunner(mode Mode, runner string) (effective, mappedFrom string, err error) {
	r := strings.TrimSpace(runner)
	if mode == "" {
		mode = ModeHeadless
	}
	switch mode {
	case ModeOpen:
		switch r {
		case "codex":
			return "codex-tty", "codex", nil
		case "grok":
			return "grok-tty", "grok", nil
		case "":
			// Leave empty: agentrunapi / prefs callers supply the default.
			return "", "", nil
		default:
			if !agenttty.IsTTYRunner(r) {
				return "", "", fmt.Errorf("--open requires a TTY runner (got %q); use codex-tty or grok-tty", r)
			}
			return r, "", nil
		}
	default: // headless
		switch r {
		case "codex-tty":
			return "codex", "codex-tty", nil
		case "grok-tty":
			return "grok", "grok-tty", nil
		case "codex", "grok", "fake-codex":
			return r, "", nil
		case "":
			return "codex", "", nil
		default:
			if agenttty.IsTTYRunner(r) {
				return "", "", fmt.Errorf("headless needs a CLI runner (got TTY %q); supported maps: codex-tty→codex, grok-tty→grok", r)
			}
			return r, "", nil
		}
	}
}
