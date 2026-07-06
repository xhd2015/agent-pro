package agenttty

import "strings"

func detectGrokScreenStatus(scrollback []byte) string {
	return detectGenericScreenStatus(scrollback, []string{"GROK_TTY_BANNER"})
}

func detectCodexScreenStatus(scrollback []byte) string {
	return detectGenericScreenStatus(scrollback, []string{"CODEX_TTY_BANNER"})
}

func detectStubScreenStatus(scrollback []byte) string {
	scenario := loadStubScenario()
	if scenario != nil && scenario.ScreenStatus != "" {
		plain := stripPlain(scrollback)
		if strings.Contains(plain, "STUB_TTY_BANNER") || strings.Contains(plain, "\u203a") {
			return scenario.ScreenStatus
		}
	}
	return detectGenericScreenStatus(scrollback, []string{"STUB_TTY_BANNER"})
}

func detectGenericScreenStatus(scrollback []byte, markers []string) string {
	if len(scrollback) == 0 {
		return "unknown"
	}
	plain := stripPlain(scrollback)
	lower := strings.ToLower(plain)
	if strings.Contains(lower, "response:") || strings.Contains(lower, "submitted:") {
		if hasPromptMarker(plain) {
			return "idle"
		}
	}
	for _, marker := range markers {
		if marker != "" && strings.Contains(plain, marker) {
			if hasPromptMarker(plain) {
				return "idle"
			}
			return "banner"
		}
	}
	if hasPromptMarker(plain) {
		return "banner"
	}
	return "unknown"
}

func hasPromptMarker(plain string) bool {
	return strings.Contains(plain, "\u203a") || strings.Contains(plain, "›") ||
		strings.Contains(plain, "❯") ||
		strings.Contains(plain, "Grok >") || strings.Contains(plain, "> ")
}

func checkGrokWritable(scrollback []byte) WritableStatus {
	plain := stripPlain(scrollback)
	if len(plain) == 0 {
		return WritableStatus{Reason: "no terminal output", State: "unknown"}
	}
	lower := strings.ToLower(plain)
	if strings.Contains(lower, "response:") || strings.Contains(lower, "submitted:") {
		return WritableStatus{Ready: true, State: "idle"}
	}
	if strings.Contains(plain, "Grok \u203a") || strings.Contains(plain, "Grok ›") ||
		strings.Contains(plain, "Grok >") || hasPromptMarker(plain) {
		if strings.Contains(lower, "working") || strings.Contains(lower, "thinking") {
			return WritableStatus{Reason: "agent still responding", State: "busy"}
		}
		return WritableStatus{Ready: true, State: "idle"}
	}
	if strings.Contains(plain, "GROK_TTY_BANNER") {
		return WritableStatus{Reason: "TUI still loading (banner not detected)", State: "loading"}
	}
	return WritableStatus{Reason: "could not detect input prompt in terminal output", State: "unknown"}
}

func checkCodexWritable(scrollback []byte) WritableStatus {
	plain := stripPlain(scrollback)
	if len(plain) == 0 {
		return WritableStatus{Reason: "no terminal output", State: "unknown"}
	}
	lower := strings.ToLower(plain)
	compact := compactWritableText(lower)
	if strings.Contains(compact, "model:loading") {
		return WritableStatus{Reason: "codex model loading", State: "loading"}
	}
	if strings.Contains(compact, "startingmcpservers") || strings.Contains(lower, "starting mcp") ||
		strings.Contains(lower, "booting mcp") {
		return WritableStatus{Reason: "codex MCP servers starting", State: "loading"}
	}
	if strings.Contains(compact, "mcpstartupincomplete") || strings.Contains(lower, "mcp startup incomplete") {
		if !strings.Contains(plain, "\u203a") && !strings.Contains(plain, "›") {
			return WritableStatus{Reason: "codex MCP startup incomplete", State: "loading"}
		}
	}
	if strings.Contains(compact, "queuedfollow-up") || strings.Contains(compact, "queuedfollowup") ||
		strings.Contains(lower, "queued follow-up") {
		return WritableStatus{Reason: "codex queued follow-up", State: "busy"}
	}
	if strings.Contains(lower, "/status") &&
		(strings.Contains(lower, "queued") || strings.Contains(lower, "follow-up") || strings.Contains(lower, "follow up")) {
		return WritableStatus{Reason: "codex /status queued", State: "busy"}
	}
	if strings.Contains(compact, "doyoutrustthecontentsofthisdirectory") {
		return WritableStatus{Reason: "codex trust prompt visible", State: "loading"}
	}
	if strings.Contains(lower, "•") && (strings.Contains(lower, "working") || strings.Contains(lower, "esc to interrupt")) {
		return WritableStatus{Reason: "codex still working (esc to interrupt)", State: "busy"}
	}
	if strings.Contains(plain, "\u203a") {
		if strings.Contains(lower, "response:") || strings.Contains(lower, "submitted:") {
			return WritableStatus{Ready: true, State: "idle"}
		}
		if !strings.Contains(lower, "working") && !strings.Contains(lower, "booting") {
			return WritableStatus{Ready: true, State: "idle"}
		}
	}
	if strings.Contains(plain, "CODEX_TTY_BANNER") {
		return WritableStatus{Reason: "TUI still loading (banner not detected)", State: "loading"}
	}
	return WritableStatus{Reason: "could not detect input prompt in terminal output", State: "unknown"}
}

func checkStubWritable(scrollback []byte) WritableStatus {
	scenario := loadStubScenario()
	plain := stripPlain(scrollback)
	lower := strings.ToLower(plain)
	if strings.Contains(lower, "working on task") ||
		(strings.Contains(plain, "•") && strings.Contains(lower, "working")) {
		return WritableStatus{Reason: "stub waiting for turn complete", State: "busy"}
	}
	if scenario != nil {
		if scenario.WritableReason != "" {
			if scenario.ScreenStatus == "busy" || scenario.ScreenStatus == "loading" {
				return WritableStatus{Reason: scenario.WritableReason, State: scenario.ScreenStatus}
			}
		}
		if scenario.ScreenStatus == "idle" {
			if strings.Contains(plain, "\u203a") || strings.Contains(plain, "STUB_TTY_BANNER") {
				return WritableStatus{Ready: true, State: "idle"}
			}
			if scenario.BannerDelayMs > 0 && !strings.Contains(plain, "STUB_TTY_BANNER") {
				return WritableStatus{Reason: "stub waiting for banner_delay_ms", State: "loading"}
			}
		}
	}
	if strings.Contains(plain, "\u203a") {
		return WritableStatus{Ready: true, State: "idle"}
	}
	if strings.Contains(plain, "STUB_TTY_BANNER") {
		return WritableStatus{Reason: "alternate screen not ready for input", State: "banner"}
	}
	return WritableStatus{Reason: "could not detect input prompt in terminal output", State: "unknown"}
}

func compactWritableText(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case ' ', '\t', '\n', '\r':
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}