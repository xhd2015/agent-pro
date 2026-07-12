package agenttty

import "strings"

// BannerDetected reports whether legacy-style banner detection matches scrollback.
// It wraps bannerDetectedConfig for characterization and fake-TUI markers.
func BannerDetected(scrollback []byte, provider string, markers []string) bool {
	return bannerDetectedConfig(scrollback, provider, markers)
}

// OpenReady reports whether grok-tty open lifecycle may proceed past banner wait.
// True for modern starting/busy/idle chrome or legacy banner markers; false for
// empty scrollback and the project-directory modal (even when legacy detection
// false-positives on the "grok build" substring).
func OpenReady(scrollback []byte) bool {
	switch ClassifyGrokScreen(scrollback) {
	case "empty", "modal":
		return false
	case "starting", "busy", "idle":
		return true
	}
	// unknown / other: accept legacy non-modal banner markers (e.g. GROK_TTY_BANNER).
	return BannerDetected(scrollback, "grok", []string{"GROK_TTY_BANNER"})
}

// ClassifyGrokScreen returns a coarse frame class:
// empty | starting | busy | idle | modal | unknown.
func ClassifyGrokScreen(scrollback []byte) string {
	plain := stripPlain(scrollback)
	if strings.TrimSpace(plain) == "" {
		return "empty"
	}
	lower := strings.ToLower(plain)

	if isGrokProjectDirectoryModal(plain, lower) {
		return "modal"
	}

	// Prefer starting over busy when session is still booting.
	if strings.Contains(lower, "starting session") && (hasModernGrokChrome(plain, lower) || hasPromptMarker(plain)) {
		return "starting"
	}

	if hasModernGrokChrome(plain, lower) {
		if grokBusyInPromptRegion(plain) || modernBusyInScreen(plain, lower) {
			return "busy"
		}
		return "idle"
	}

	// Legacy Grok › / response-style chrome.
	if strings.Contains(plain, "Grok ›") || strings.Contains(plain, "Grok \u203a") ||
		(strings.Contains(lower, "grok") && strings.Contains(plain, "›")) ||
		strings.Contains(plain, "Grok >") {
		if grokBusyInPromptRegion(plain) {
			return "busy"
		}
		return "idle"
	}

	// Generic prompt with busy / idle signals (historical fixtures).
	if hasPromptMarker(plain) {
		if grokBusyInPromptRegion(plain) {
			return "busy"
		}
		return "idle"
	}

	return "unknown"
}

func isGrokProvider(provider string) bool {
	return provider == "grok" || provider == "grok-tty"
}

func isGrokProjectDirectoryModal(plain, lower string) bool {
	if strings.Contains(lower, "run grok build in a project directory") {
		return true
	}
	// Secondary markers for the workspace picker (Enter:submit radio UI).
	if strings.Contains(lower, "type your answer here") &&
		(strings.Contains(plain, "Enter:submit") || strings.Contains(plain, "(○)") || strings.Contains(plain, "(o)")) {
		return true
	}
	return false
}

// hasModernGrokChrome reports modern SeaTalk-era TUI chrome (❯ + model/footer/status).
func hasModernGrokChrome(plain, lower string) bool {
	if !strings.Contains(plain, "❯") {
		return false
	}
	return strings.Contains(plain, "Grok 4.5") ||
		strings.Contains(plain, "Shift+Tab:mode") ||
		strings.Contains(lower, "starting session") ||
		strings.Contains(lower, "thinking") ||
		strings.Contains(lower, "tasks") ||
		strings.Contains(lower, "turn completed") ||
		strings.Contains(plain, "always-approve")
}

// modernBusyInScreen detects modern busy chrome (Thinking… / Tasks) near the
// prompt region without relying on full-scrollback "tasks" false positives.
func modernBusyInScreen(plain, lower string) bool {
	region := strings.ToLower(grokPromptRegion(plain))
	if strings.Contains(region, "thinking") {
		return true
	}
	// Tasks panel often sits above the input box while the agent is working.
	if strings.Contains(region, "tasks") && strings.Contains(region, "thinking") {
		return true
	}
	// Fallback: spinner-style "Thinking…" anywhere with modern chrome present.
	return strings.Contains(lower, "thinking…") || strings.Contains(lower, "thinking...")
}
