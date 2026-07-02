package groktty

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hinshun/vt10x"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

const (
	testBannerMarker       = "GROK_TTY_BANNER"
	codexTestBannerMarker  = "CODEX_TTY_BANNER"
	bannerWaitTimeout      = 30 * time.Second
	codexBannerWaitTimeout = 2 * time.Minute
	bannerPollInterval     = 75 * time.Millisecond
)

var (
	responseLineRe  = regexp.MustCompile(`(?m)Response:\s*(.+)`)
	submittedLineRe = regexp.MustCompile(`(?m)SUBMITTED:(.+)`)
)

func waitForBanner(ctx context.Context, mgr *ptywrap.Manager, sessionID string) error {
	return waitForBannerConfig(ctx, mgr, sessionID, "grok", []string{testBannerMarker})
}

func waitForBannerConfig(ctx context.Context, mgr *ptywrap.Manager, sessionID, provider string, markers []string) error {
	waitTimeout := bannerWaitTimeout
	if provider == "codex" || provider == "codex-tty" {
		waitTimeout = codexBannerWaitTimeout
	}
	deadline := time.Now().Add(waitTimeout)
	trustPromptAccepted := false
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		scrollback := mgr.Scrollback(sessionID)
		if codexTrustPromptDetected(scrollback, provider) && !trustPromptAccepted {
			trustPromptAccepted = true
			if err := mgr.WriteInput(sessionID, []byte("\r")); err != nil {
				return err
			}
			time.Sleep(bannerPollInterval)
			continue
		}
		if bannerDetectedConfig(scrollback, provider, markers) {
			return nil
		}
		if isCodexProvider(provider) && bannerDetectedConfig(renderTerminalText(scrollback), provider, nil) {
			return nil
		}
		time.Sleep(bannerPollInterval)
	}
	return fmt.Errorf("%s TUI banner not detected", provider)
}

func bannerDetected(scrollback []byte) bool {
	return bannerDetectedConfig(scrollback, "grok", []string{testBannerMarker})
}

func bannerDetectedConfig(scrollback []byte, provider string, markers []string) bool {
	plain := stripANSI(scrollback)
	for _, marker := range markers {
		if marker != "" && strings.Contains(plain, marker) {
			return true
		}
	}
	lower := strings.ToLower(plain)
	if isCodexProvider(provider) {
		compact := compactBannerText(lower)
		if codexModelLoadingScreen(compact) {
			return false
		}
		if strings.Contains(plain, "Codex ›") || strings.Contains(plain, "Codex \u203a") {
			return true
		}
		if strings.Contains(lower, "codex") && strings.Contains(plain, "›") {
			return true
		}
		if strings.Contains(lower, "codex") && strings.Contains(lower, "ready") {
			return true
		}
		if strings.Contains(compact, "openaicodex") &&
			strings.Contains(plain, "›") &&
			(strings.Contains(compact, "writetestsfor@filename") ||
				(strings.Contains(compact, "model:") && strings.Contains(compact, "directory:"))) &&
			!strings.Contains(compact, "bootingmcpserver") &&
			!strings.Contains(compact, "startingmcpservers") {
			return true
		}
		if strings.Contains(plain, "›") &&
			strings.Contains(compact, "writetestsfor@filename") &&
			strings.Contains(compact, "gpt-") {
			return true
		}
		return false
	}
	if strings.Contains(plain, "Grok ›") || strings.Contains(plain, "Grok \u203a") {
		return true
	}
	if strings.Contains(lower, "grok") && strings.Contains(plain, "›") {
		return true
	}
	if strings.Contains(lower, "grok build") {
		return true
	}
	return false
}

func codexModelLoadingScreen(compact string) bool {
	return strings.Contains(compact, "openaicodex") &&
		strings.Contains(compact, "model:loading") &&
		strings.Contains(compact, "/modeltochange")
}

func codexTrustPromptDetected(scrollback []byte, provider string) bool {
	if !isCodexProvider(provider) {
		return false
	}
	compact := compactBannerText(strings.ToLower(stripANSI(scrollback)))
	return strings.Contains(compact, "doyoutrustthecontentsofthisdirectory") &&
		strings.Contains(compact, "pressentertocontinue")
}

func autoExitCodexAfterTurn(ctx context.Context, mgr *ptywrap.Manager, sessionID, prompt string) {
	promptCompact := compactBannerText(strings.ToLower(prompt))
	if promptCompact == "" {
		return
	}
	minReadyAt := time.Now().Add(2 * time.Second)
	transportTimeoutExitAt := time.Now().Add(10 * time.Second)
	stuckExitAt := time.Now().Add(90 * time.Second)
	var completeSince time.Time
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if time.Now().Before(minReadyAt) {
			continue
		}
		if codexTurnCompleteForExit(mgr.Scrollback(sessionID), promptCompact) {
			if completeSince.IsZero() {
				completeSince = time.Now()
				continue
			}
			if time.Since(completeSince) >= 10*time.Second {
				_ = mgr.WriteInput(sessionID, []byte{0x04})
				return
			}
		} else {
			completeSince = time.Time{}
		}
		scrollback := mgr.Scrollback(sessionID)
		if !time.Now().Before(transportTimeoutExitAt) && codexTransportTimedOut(scrollback) {
			_ = mgr.WriteInput(sessionID, []byte{0x04})
			return
		}
		if !time.Now().Before(stuckExitAt) {
			_ = mgr.WriteInput(sessionID, []byte{0x04})
			return
		}
	}
}

func waitForPersistentTurn(ctx context.Context, mgr *ptywrap.Manager, sessionID, prompt string, cfg runConfig) error {
	timeout := 90 * time.Second
	if isCodexProvider(cfg.bannerProvider) || cfg.runnerID == "codex-tty" {
		timeout = 3 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	promptCompact := compactBannerText(strings.ToLower(prompt))
	var completeSince time.Time
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		scrollback := mgr.Scrollback(sessionID)
		if persistentTurnComplete(scrollback, prompt, promptCompact, cfg) {
			if completeSince.IsZero() {
				completeSince = time.Now()
				continue
			}
			if time.Since(completeSince) >= 750*time.Millisecond {
				return nil
			}
			continue
		}
		completeSince = time.Time{}
	}
}

func persistentTurnComplete(scrollback []byte, prompt, promptCompact string, cfg runConfig) bool {
	if isCodexProvider(cfg.bannerProvider) || cfg.runnerID == "codex-tty" {
		if strings.TrimSpace(extractAssistantTextConfigForProvider(scrollback, prompt, cfg.bannerMarkers, cfg.bannerProvider)) != "" {
			return true
		}
		return codexTurnCompleteForExit(scrollback, promptCompact)
	}
	return strings.TrimSpace(extractAssistantTextConfigForProvider(scrollback, prompt, cfg.bannerMarkers, cfg.bannerProvider)) != ""
}

func retryCodexSubmit(ctx context.Context, mgr *ptywrap.Manager, sessionID, prompt string) {
	promptCompact := compactBannerText(strings.ToLower(prompt))
	if promptCompact == "" {
		return
	}
	ticker := time.NewTicker(750 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.Now().Add(2 * time.Minute)
	var lastQueueAttempt time.Time
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		scrollback := mgr.Scrollback(sessionID)
		if codexTurnCompleteForExit(scrollback, promptCompact) {
			return
		}
		if !codexPromptStillInInput(scrollback, promptCompact) {
			if codexPromptVisibleInInput(scrollback, promptCompact) && time.Since(lastQueueAttempt) >= 2*time.Second {
				lastQueueAttempt = time.Now()
				if err := mgr.WriteInput(sessionID, []byte("\t")); err != nil {
					return
				}
			}
			continue
		}
		if err := mgr.WriteInput(sessionID, []byte("\r")); err != nil {
			return
		}
	}
}

func codexPromptStillInInput(scrollback []byte, promptCompact string) bool {
	plain := stripANSI(renderTerminalText(scrollback))
	if strings.Contains(plain, "•") {
		return false
	}
	compact := compactBannerText(strings.ToLower(plain))
	if !strings.Contains(compact, "›"+promptCompact) {
		return false
	}
	if codexModelLoadingScreen(compact) {
		return false
	}
	for _, marker := range codexBusyMarkers() {
		if strings.Contains(compact, marker) {
			return false
		}
	}
	return true
}

func codexPromptVisibleInInput(scrollback []byte, promptCompact string) bool {
	if promptCompact == "" {
		return false
	}
	rawCompact := compactBannerText(strings.ToLower(stripANSI(scrollback)))
	screenCompact := compactBannerText(strings.ToLower(stripANSI(renderTerminalText(scrollback))))
	return strings.Contains(rawCompact, "›"+promptCompact) ||
		strings.Contains(screenCompact, "›"+promptCompact)
}

func codexTurnCompleteForExit(scrollback []byte, promptCompact string) bool {
	rawPlain := stripANSI(scrollback)
	if !strings.Contains(rawPlain, "•") && codexPromptStillInInput(scrollback, promptCompact) {
		return false
	}
	screenPlain := stripANSI(renderTerminalText(scrollback))
	if !strings.Contains(compactBannerText(strings.ToLower(rawPlain)), promptCompact) &&
		!strings.Contains(compactBannerText(strings.ToLower(screenPlain)), promptCompact) {
		return false
	}
	if strings.TrimSpace(screenPlain) == "" {
		screenPlain = rawPlain
	}
	return codexTurnCompleteText(screenPlain, promptCompact)
}

func codexTurnCompleteText(plain string, promptCompact string) bool {
	compact := compactBannerText(strings.ToLower(plain))
	if !strings.Contains(plain, "›") {
		return false
	}
	for _, marker := range codexBusyMarkers() {
		if strings.Contains(compact, marker) {
			return false
		}
	}
	if codexModelLoadingScreen(compact) {
		return false
	}
	if strings.Contains(compact, "openaicodex") ||
		strings.Contains(compact, "gpt-") {
		return true
	}
	return false
}

func codexBusyMarkers() []string {
	return []string{
		"working",
		"bootingmcpserver",
		"startingmcpservers",
		"runningstophook",
		"esctointerrupt",
		"queuedfollow-upinputs",
	}
}

func codexTransportTimedOut(scrollback []byte) bool {
	plain := stripANSI(renderTerminalText(scrollback))
	compact := compactBannerText(strings.ToLower(plain))
	return strings.Contains(compact, "requesttimedout") &&
		strings.Contains(compact, "working")
}

func isCodexProvider(provider string) bool {
	return provider == "codex" || provider == "codex-tty"
}

func renderTerminalText(scrollback []byte) []byte {
	if len(scrollback) == 0 {
		return nil
	}
	vt := vt10x.New(vt10x.WithSize(80, 24))
	if _, err := vt.Write(scrollback); err != nil {
		return nil
	}
	vt.Lock()
	defer vt.Unlock()

	var out strings.Builder
	for y := 0; y < 24; y++ {
		line := renderTerminalLine(vt, 80, y)
		if line == "" {
			continue
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return []byte(out.String())
}

func renderTerminalLine(vt vt10x.Terminal, cols, y int) string {
	runes := make([]rune, cols)
	lastNonSpace := -1
	for x := 0; x < cols; x++ {
		ch := vt.Cell(x, y).Char
		if ch == 0 {
			ch = ' '
		}
		runes[x] = ch
		if ch != ' ' {
			lastNonSpace = x
		}
	}
	if lastNonSpace < 0 {
		return ""
	}
	return string(runes[:lastNonSpace+1])
}

func compactBannerText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\b', '\f', '\n', '\r', '\t', ' ', '`', '\'', '"':
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func extractAssistantText(scrollback []byte, prompt string) string {
	return extractAssistantTextConfig(scrollback, prompt, []string{testBannerMarker})
}

func extractAssistantTextConfig(scrollback []byte, prompt string, markers []string) string {
	return extractAssistantTextConfigForProvider(scrollback, prompt, markers, "")
}

func extractAssistantTextConfigForProvider(scrollback []byte, prompt string, markers []string, provider string) string {
	plain := stripANSI(scrollback)
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
	plain := strings.TrimSpace(stripANSI(renderTerminalText(scrollback)))
	if plain == "" {
		plain = stripANSI(scrollback)
	}
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
