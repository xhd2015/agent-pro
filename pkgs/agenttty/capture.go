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
		if bulletText := extractCodexBulletText(line, prompt, markers); bulletText != "" {
			kept = append(kept, bulletText)
			continue
		}
		if skipCodexFallbackLine(line, prompt, markers) {
			continue
		}
		kept = append(kept, line)
	}
	if len(kept) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
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