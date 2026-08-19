package agenttty

import (
	"regexp"
	"strings"
)

var (
	responseLineRe  = regexp.MustCompile(`(?m)Response:\s*(.+)`)
	submittedLineRe = regexp.MustCompile(`(?m)SUBMITTED:(.+)`)
)

// ExtractAssistantTextFromSnapshot extracts assistant response text from terminal scrollback.
func ExtractAssistantTextFromSnapshot(runner string, scrollback []byte, prompt string) string {
	provider, ok := Get(runner)
	markers := []string(nil)
	bannerProvider := runner
	if ok {
		markers = provider.BannerMarkers
		bannerProvider = provider.BannerProvider
	}
	return extractAssistantTextForProvider(scrollback, prompt, markers, bannerProvider)
}

func extractAssistantTextForProvider(scrollback []byte, prompt string, markers []string, provider string) string {
	plain := stripPlain(scrollback)
	if matches := submittedLineRe.FindStringSubmatch(plain); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	if matches := responseLineRe.FindStringSubmatch(plain); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	if isCodexProvider(provider) {
		return cleanCodexScrollbackFallback(scrollback, prompt, markers)
	}
	if provider == "commandcode" {
		return cleanCommandcodeScrollback(scrollback, prompt)
	}

	lines := strings.Split(plain, "\n")
	var kept []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		skip := false
		for _, marker := range markers {
			if marker != "" && strings.Contains(line, marker) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if strings.HasPrefix(line, "Grok") && strings.Contains(line, "›") {
			continue
		}
		if strings.HasPrefix(line, "Codex") && strings.Contains(line, "›") {
			continue
		}
		if strings.EqualFold(line, prompt) {
			continue
		}
		if strings.Contains(line, "[Terminal exited]") {
			continue
		}
		kept = append(kept, line)
	}
	if len(kept) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

func cleanCodexScrollbackFallback(scrollback []byte, prompt string, markers []string) string {
	plain := strings.TrimSpace(stripPlain(scrollback))
	lines := strings.Split(plain, "\n")
	var kept []string
	for _, line := range lines {
		line = cleanTerminalTextLine(line)
		if line == "" {
			continue
		}
		// PTY snapshots sometimes glue the prompt row to the next printf
		// ("› run lsls output:"). Recover useful text after the prompt glyph.
		if i := strings.Index(line, "›"); i >= 0 {
			if j := strings.Index(line, "ls output:"); j >= 0 {
				line = cleanTerminalTextLine(line[j:])
			} else {
				continue
			}
		}
		if bulletText := extractCodexBulletText(line, prompt, markers); bulletText != "" {
			kept = append(kept, bulletText)
			continue
		}
		if skipCodexFallbackLine(line, prompt, markers) {
			continue
		}
		kept = append(kept, splitGluedCodexResultLines(line)...)
	}
	if len(kept) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

// splitGluedCodexResultLines repairs snapshots that dropped LFs between short
// result rows ("AGENTS.mdcmdpkgs" → separate lines).
func splitGluedCodexResultLines(line string) []string {
	tokens := []string{"ls output:", "AGENTS.md", "cmd", "pkgs"}
	if !strings.Contains(line, "AGENTS.md") || !strings.Contains(line, "pkgs") {
		return []string{line}
	}
	var out []string
	rest := line
	for _, tok := range tokens {
		if i := strings.Index(rest, tok); i >= 0 {
			if prefix := strings.TrimSpace(rest[:i]); prefix != "" {
				out = append(out, prefix)
			}
			out = append(out, tok)
			rest = rest[i+len(tok):]
		}
	}
	if tail := strings.TrimSpace(rest); tail != "" {
		out = append(out, tail)
	}
	if len(out) == 0 {
		return []string{line}
	}
	return out
}

func extractCodexBulletText(line, prompt string, markers []string) string {
	if !strings.Contains(line, "•") {
		return ""
	}
	var kept []string
	for _, segment := range strings.Split(line, "•")[1:] {
		if idx := strings.Index(segment, "›"); idx >= 0 {
			segment = segment[:idx]
		}
		segment = cleanTerminalTextLine(segment)
		if segment == "" || skipCodexFallbackLine(segment, prompt, markers) {
			continue
		}
		lower := strings.ToLower(segment)
		if strings.HasPrefix(lower, "working") ||
			strings.HasPrefix(lower, "running ") ||
			strings.HasPrefix(lower, "starting ") ||
			strings.HasPrefix(lower, "queued ") ||
			strings.Contains(lower, "esc to interrupt") {
			continue
		}
		kept = append(kept, segment)
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

func cleanTerminalTextLine(line string) string {
	line = strings.TrimSpace(line)
	return strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\t' {
			return -1
		}
		return r
	}, line)
}

// isASCIIRuleLine reports separator rows that are only dashes/equals (box
// drawing often flattened to '-' in PTY snapshots).
func isASCIIRuleLine(line string) bool {
	if line == "" {
		return false
	}
	for _, r := range line {
		if r != '-' && r != '=' && r != '─' && r != '═' && r != ' ' && r != '\t' {
			return false
		}
	}
	return strings.ContainsAny(line, "-=─═")
}

func skipCodexFallbackLine(line, prompt string, markers []string) bool {
	for _, marker := range markers {
		if marker != "" && strings.Contains(line, marker) {
			return true
		}
	}
	if strings.EqualFold(line, strings.TrimSpace(prompt)) {
		return true
	}
	if strings.Contains(line, "[Terminal exited]") {
		return true
	}
	if strings.ContainsAny(line, "╭╮╰╯│─") {
		return true
	}
	// PTY snapshots often flatten box-drawing horizontals to ASCII '-'.
	if isASCIIRuleLine(line) {
		return true
	}
	lower := strings.ToLower(line)
	compact := compactBannerText(lower)
	if strings.Contains(line, "›") {
		return true
	}
	if strings.Contains(line, ">4;0m") || strings.Contains(line, ">7u") {
		return true
	}
	if strings.Contains(lower, "openai codex") ||
		strings.Contains(lower, "[features].codex_hooks") ||
		strings.Contains(lower, "[features].hooks") ||
		strings.Contains(lower, "developers.openai.com/codex") ||
		strings.HasPrefix(lower, "enable it with") ||
		strings.HasPrefix(lower, "for details") ||
		strings.HasPrefix(lower, "tip:") ||
		strings.HasPrefix(lower, "permissions:") ||
		strings.Contains(compact, "model:loading") ||
		strings.HasPrefix(lower, "model:") ||
		strings.HasPrefix(lower, "directory:") ||
		strings.Contains(lower, "starting mcp servers") ||
		strings.Contains(lower, "booting mcp") ||
		strings.Contains(lower, "running stop hook") ||
		strings.Contains(lower, "running userpromptsubmit hook") ||
		strings.HasPrefix(lower, "working") {
		return true
	}
	return false
}

func cleanCommandcodeScrollback(scrollback []byte, prompt string) string {
	plain := strings.TrimSpace(stripPlain(scrollback))
	lines := strings.Split(plain, "\n")
	var kept []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if skipCommandcodeTuiLine(line, prompt) {
			continue
		}
		kept = append(kept, line)
	}
	if len(kept) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

func skipCommandcodeTuiLine(line, prompt string) bool {
	lower := strings.ToLower(line)
	if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "#") {
		return true
	}
	if strings.Contains(line, "\u2588") {
		return true
	}
	if strings.Contains(line, "v0.") && strings.Contains(lower, "command") {
		return true
	}
	if strings.Trim(line, "\u2500\u2014\u2015") == "" {
		return true
	}
	if strings.TrimSpace(line) == strings.TrimRight(strings.TrimSpace(line), "\u2500") {
		dashes := strings.Count(line, "\u2500")
		if dashes >= 10 {
			return true
		}
	}
	switch {
	case strings.HasPrefix(line, "\u203a") && strings.Contains(line, "\u2039"):
		return true
	case strings.HasPrefix(line, "\u203a ") || strings.HasPrefix(line, "❯ "):
		return true
	case strings.HasPrefix(lower, "\u203a ask") || strings.Contains(lower, "ask your question"):
		return true
	case strings.HasPrefix(line, "\u00bb") || strings.HasPrefix(line, "»"):
		return true
	case strings.HasPrefix(line, "? ") || strings.HasPrefix(line, "\\u003f "):
		return true
	}
	if strings.EqualFold(strings.TrimPrefix(line, "\u203a "), prompt) {
		return true
	}
	if strings.EqualFold(line, prompt) {
		return true
	}
	if strings.Contains(line, "\u2022") {
		if strings.Contains(lower, "esc") || strings.Contains(lower, "interrupt") ||
			strings.Contains(lower, "ctrl+t") || strings.Contains(lower, "taste") ||
			strings.Contains(lower, "shift+tab") || strings.Contains(lower, "reviewing") ||
			strings.Contains(lower, "concept") || strings.Contains(lower, "bypass") {
			return true
		}
	}
	if strings.Contains(lower, "retrying") || strings.Contains(lower, "connection issue") || strings.Contains(lower, "attempt") {
		return true
	}
	if strings.Contains(lower, "permission bypass") || strings.Contains(lower, "for shortcuts") {
		return true
	}
	if strings.HasPrefix(line, "[") && strings.Contains(line, "Terminal exited") {
		return true
	}
	// Horizontal separator lines (all same repeating char).
	if len(line) >= 10 {
		allSame := true
		for _, r := range line {
			if r != rune(line[0]) {
				allSame = false
				break
			}
		}
		if allSame && (line[0] == '-' || line[0] == '_' || line[0] == '=') {
			return true
		}
	}
	return false
}
