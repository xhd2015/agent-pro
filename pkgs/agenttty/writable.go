package agenttty

import "strings"

func detectGrokScreenStatus(scrollback []byte) string {
	// Align tty status / idle watchdog with ClassifyGrokScreen (and writable
	// Ready). Generic banner markers never appear on live Grok chrome.
	switch class := ClassifyGrokScreen(scrollback); class {
	case "idle", "busy", "starting", "modal":
		return class
	case "empty":
		return "unknown"
	}
	st := checkGrokWritable(scrollback)
	if st.Ready {
		return "idle"
	}
	if s := strings.TrimSpace(st.State); s != "" && s != "unknown" {
		return s
	}
	return "unknown"
}

func detectCodexScreenStatus(scrollback []byte) string {
	// Align tty status / idle watchdog with checkCodexWritable. Live Codex
	// never prints CODEX_TTY_BANNER, so the generic stub detector would
	// leave a finished ›/» prompt at "banner" and SampleIsIdle never arms.
	st := checkCodexWritable(scrollback)
	if st.Ready {
		return "idle"
	}
	if s := strings.TrimSpace(st.State); s != "" && s != "unknown" {
		return s
	}
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

// hasCodexPromptMarker reports main-chat composer glyphs used by Codex:
// legacy › (U+203A) and Codex 0.146+ » (U+00BB).
func hasCodexPromptMarker(plain string) bool {
	return strings.Contains(plain, "\u203a") || strings.Contains(plain, "›") ||
		strings.Contains(plain, "\u00bb") || strings.Contains(plain, "»")
}

func hasPromptMarker(plain string) bool {
	return hasCodexPromptMarker(plain) ||
		strings.Contains(plain, "❯") ||
		strings.Contains(plain, "Grok >") || strings.Contains(plain, "> ")
}

func grokPromptRegion(full string) string {
	markers := []string{"Enter:send", "Shift+Tab:mode", "Ctrl+.:shortcuts", "╭---"}
	best := 0
	for _, m := range markers {
		if i := strings.LastIndex(full, m); i > best {
			best = i
		}
	}
	if best > 0 {
		start := best
		if start > 400 {
			start -= 400
		} else {
			start = 0
		}
		return full[start:]
	}
	return grokTailLines(full, 16)
}

func grokTailLines(text string, n int) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= n {
		return text
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}

func grokBusyInPromptRegion(plain string) bool {
	region := strings.ToLower(grokPromptRegion(plain))
	region = strings.ReplaceAll(region, "working tree", "wt_scrubbed")
	return strings.Contains(region, "working") || strings.Contains(region, "thinking")
}

// codexPromptRegion returns the active/post-turn slice of Codex scrollback used for
// busy vs idle. Prefer the last settled "Worked for" footer (everything above is
// historical). Otherwise prefer context around the last main-chat ›/» glyph.
// Falls back to a short tail so live "• Working" without a prompt still matches.
func codexPromptRegion(plain string) string {
	lower := strings.ToLower(plain)
	if i := strings.LastIndex(lower, "worked for"); i >= 0 {
		return plain[i:]
	}
	best := -1
	for _, m := range []string{"\u203a", "›", "\u00bb", "»"} {
		if j := strings.LastIndex(plain, m); j > best {
			best = j
		}
	}
	if best >= 0 {
		// Include lines above the prompt so live Working chrome just above › still counts.
		start := best
		if start > 500 {
			start -= 500
		} else {
			start = 0
		}
		// Snap back to a line start when possible.
		if nl := strings.LastIndex(plain[start:best], "\n"); nl >= 0 {
			start = start + nl + 1
		}
		return plain[start:]
	}
	return grokTailLines(plain, 16)
}

// codexBusyInPromptRegion reports live Codex busy chrome (• + working / esc to interrupt)
// only inside codexPromptRegion — not historical markers above a settled post-turn footer.
func codexBusyInPromptRegion(plain string) bool {
	region := strings.ToLower(codexPromptRegion(plain))
	hasBullet := strings.Contains(region, "•") || strings.Contains(region, "\u2022")
	if !hasBullet {
		return false
	}
	return strings.Contains(region, "working") || strings.Contains(region, "esc to interrupt")
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
		if grokBusyInPromptRegion(plain) {
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
	// After /exit the keep-alive scrollback still holds historical › glyphs.
	// Exit footer (phrase ∧ codex resume) or [Terminal exited] ⇒ not injectable.
	if TerminalExitedMarkerPresent(plain) {
		return WritableStatus{Reason: "terminal exited", State: "exited"}
	}
	if CodexExitFooterPresent(plain) {
		return WritableStatus{Reason: "codex agent exited (resume footer)", State: "exited"}
	}
	lower := strings.ToLower(plain)
	compact := compactWritableText(lower)
	// Only the blocking Update available *menu* is non-writable. Residual banners that
	// still say "Update available" / "Run npm install" after Skip must not hang waitForPrompt.
	if IsBlockingUpdateMenu(plain) {
		return WritableStatus{Reason: "codex update available", State: "loading"}
	}
	if strings.Contains(compact, "model:loading") {
		return WritableStatus{Reason: "codex model loading", State: "loading"}
	}
	if strings.Contains(compact, "startingmcpservers") || strings.Contains(lower, "starting mcp") ||
		strings.Contains(lower, "booting mcp") {
		return WritableStatus{Reason: "codex MCP servers starting", State: "loading"}
	}
	if strings.Contains(compact, "mcpstartupincomplete") || strings.Contains(lower, "mcp startup incomplete") {
		if !hasCodexPromptMarker(plain) {
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
	// Busy only in the active/post-turn region — historical "• Working" above a settled
	// "Worked for …" + bottom › must not block WaitDone/send (scorer idle false-negative).
	if codexBusyInPromptRegion(plain) {
		return WritableStatus{Reason: "codex still working (esc to interrupt)", State: "busy"}
	}
	if hasCodexPromptMarker(plain) {
		if strings.Contains(lower, "response:") || strings.Contains(lower, "submitted:") {
			return WritableStatus{Ready: true, State: "idle"}
		}
		// "working"/"booting" only matter in the prompt region (same as busy rule).
		regionLower := strings.ToLower(codexPromptRegion(plain))
		if !strings.Contains(regionLower, "working") && !strings.Contains(regionLower, "booting") {
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
	if strings.Contains(plain, "\u203a") || strings.Contains(plain, "›") {
		return WritableStatus{Ready: true, State: "idle"}
	}
	if scenario != nil {
		if scenario.WritableReason != "" {
			if scenario.ScreenStatus == "busy" || scenario.ScreenStatus == "loading" {
				return WritableStatus{Reason: scenario.WritableReason, State: scenario.ScreenStatus}
			}
		}
		if scenario.ScreenStatus == "idle" {
			if strings.Contains(plain, "STUB_TTY_BANNER") {
				return WritableStatus{Ready: true, State: "idle"}
			}
			if scenario.BannerDelayMs > 0 && !strings.Contains(plain, "STUB_TTY_BANNER") {
				return WritableStatus{Reason: "stub waiting for banner_delay_ms", State: "loading"}
			}
		}
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