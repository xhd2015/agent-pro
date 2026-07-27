package agenttty

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

const (
	bannerWaitTimeout      = 30 * time.Second
	codexBannerWaitTimeout = 2 * time.Minute
	bannerPollInterval     = 75 * time.Millisecond
)

type runConfig struct {
	runnerID       string
	bannerProvider string
	bannerMarkers  []string
}

func waitForBannerRemote(ctx context.Context, listenAddr, sessionID, provider string, markers []string) error {
	waitTimeout := bannerWaitTimeout
	if isCodexProvider(provider) {
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
		snapshot, err := ttywatch.SnapshotText(listenAddr, sessionID)
		if err != nil {
			time.Sleep(bannerPollInterval)
			continue
		}
		scrollback := []byte(snapshot)
		if codexTrustPromptDetected(scrollback, provider) && !trustPromptAccepted {
			trustPromptAccepted = true
			if err := ttywatch.SendMessage(listenAddr, sessionID, "", true); err != nil {
				return err
			}
			time.Sleep(bannerPollInterval)
			continue
		}
		// Grok: use OpenReady so modern starting/busy/idle chrome succeeds and the
		// project-directory modal does not (legacy "grok build" false-positive).
		if isGrokProvider(provider) {
			if OpenReady(scrollback) {
				return nil
			}
		} else if bannerDetectedConfig(scrollback, provider, markers) {
			return nil
		}
		time.Sleep(bannerPollInterval)
	}
	return fmt.Errorf("%s TUI banner not detected", provider)
}

// acceptCodexTrustRemote soft-polls for the directory trust modal and sends Enter.
// Used by --open (which does not hard-wait on banner) so sendable is not blocked
// by trust or a false "update available" classification of the enter footer.
// Returns when trust is gone or timeout/ctx elapses (never hard-fails open).
func acceptCodexTrustRemote(ctx context.Context, listenAddr, sessionID, provider string, timeout time.Duration) {
	if !isCodexProvider(provider) {
		return
	}
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	deadline := time.Now().Add(timeout)
	accepted := false
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return
		default:
		}
		snapshot, err := ttywatch.SnapshotText(listenAddr, sessionID)
		if err != nil {
			time.Sleep(bannerPollInterval)
			continue
		}
		scrollback := []byte(snapshot)
		if codexTrustPromptDetected(scrollback, provider) {
			if !accepted {
				_ = ttywatch.SendMessage(listenAddr, sessionID, "", true)
				accepted = true
			}
			time.Sleep(bannerPollInterval)
			continue
		}
		// Trust cleared; optionally wait briefly for idle chrome.
		plain := stripPlain(scrollback)
		if strings.Contains(plain, "›") || strings.Contains(plain, "\u203a") ||
			strings.Contains(plain, "OpenAI Codex") || bannerDetectedConfig(scrollback, provider, nil) {
			return
		}
		time.Sleep(bannerPollInterval)
	}
}

func bannerDetectedConfig(scrollback []byte, provider string, markers []string) bool {
	plain := stripPlain(scrollback)
	if provider == "stub" {
		if strings.Contains(plain, "\u203a") || strings.Contains(plain, "›") {
			return true
		}
	}
	if provider == "commandcode" {
		return strings.TrimSpace(plain) != ""
	}
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
	compact := compactBannerText(strings.ToLower(stripPlain(scrollback)))
	return strings.Contains(compact, "doyoutrustthecontentsofthisdirectory") &&
		strings.Contains(compact, "pressentertocontinue")
}

func autoExitCodexAfterTurnRemote(ctx context.Context, listenAddr, sessionID, prompt string) {
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
		snapshot, err := fetchSnapshotBytes(listenAddr, sessionID)
		if err != nil {
			continue
		}
		if codexTurnCompleteForExit(snapshot, promptCompact) {
			if completeSince.IsZero() {
				completeSince = time.Now()
				continue
			}
			if time.Since(completeSince) >= 10*time.Second {
				_ = ttywatch.InjectInput(listenAddr, sessionID, []byte{0x04})
				return
			}
		} else {
			completeSince = time.Time{}
		}
		if !time.Now().Before(transportTimeoutExitAt) && codexTransportTimedOut(snapshot) {
			_ = ttywatch.InjectInput(listenAddr, sessionID, []byte{0x04})
			return
		}
		if !time.Now().Before(stuckExitAt) {
			_ = ttywatch.InjectInput(listenAddr, sessionID, []byte{0x04})
			return
		}
	}
}

func waitForPersistentTurnRemote(ctx context.Context, listenAddr, sessionID, prompt string, cfg runConfig, extraComplete func() bool) error {
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
		complete := false
		if extraComplete != nil && cfg.runnerID == "grok-tty" {
			complete = extraComplete()
		} else {
			snapshot, err := fetchSnapshotBytes(listenAddr, sessionID)
			if err == nil {
				complete = persistentTurnComplete(snapshot, prompt, promptCompact, cfg)
			}
			if !complete && extraComplete != nil {
				complete = extraComplete()
			}
		}
		if complete {
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
		if strings.TrimSpace(extractAssistantTextForProvider(scrollback, prompt, cfg.bannerMarkers, cfg.bannerProvider)) != "" {
			return true
		}
		return codexTurnCompleteForExit(scrollback, promptCompact)
	}
	if cfg.runnerID == "commandcode-tty" {
		// commandcode-tty: headless uses -p which exits cleanly.
		// Wait for scrollback content stability rather than text length.
		// The keep-alive serve stays up but we exit the wait loop here.
		captured := strings.TrimSpace(extractAssistantTextForProvider(scrollback, prompt, cfg.bannerMarkers, cfg.bannerProvider))
		return captured != ""
	}
	return strings.TrimSpace(extractAssistantTextForProvider(scrollback, prompt, cfg.bannerMarkers, cfg.bannerProvider)) != ""
}

func retryCodexSubmitRemote(ctx context.Context, listenAddr, sessionID, prompt string) {
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
		snapshot, err := fetchSnapshotBytes(listenAddr, sessionID)
		if err != nil {
			continue
		}
		if codexTurnCompleteForExit(snapshot, promptCompact) {
			return
		}
		if !codexPromptStillInInput(snapshot, promptCompact) {
			if codexPromptVisibleInInput(snapshot, promptCompact) && time.Since(lastQueueAttempt) >= 2*time.Second {
				lastQueueAttempt = time.Now()
				if err := ttywatch.InjectInput(listenAddr, sessionID, []byte("\t")); err != nil {
					return
				}
			}
			continue
		}
		if err := ttywatch.SendMessage(listenAddr, sessionID, "", true); err != nil {
			return
		}
	}
}

func codexPromptStillInInput(scrollback []byte, promptCompact string) bool {
	plain := stripPlain(scrollback)
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
	compact := compactBannerText(strings.ToLower(stripPlain(scrollback)))
	return strings.Contains(compact, "›"+promptCompact)
}

func codexTurnCompleteForExit(scrollback []byte, promptCompact string) bool {
	rawPlain := stripPlain(scrollback)
	if !strings.Contains(rawPlain, "•") && codexPromptStillInInput(scrollback, promptCompact) {
		return false
	}
	screenPlain := rawPlain
	if !strings.Contains(compactBannerText(strings.ToLower(rawPlain)), promptCompact) &&
		!strings.Contains(compactBannerText(strings.ToLower(screenPlain)), promptCompact) {
		return false
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
	plain := stripPlain(scrollback)
	compact := compactBannerText(strings.ToLower(plain))
	return strings.Contains(compact, "requesttimedout") &&
		strings.Contains(compact, "working")
}

func isCodexProvider(provider string) bool {
	return provider == "codex" || provider == "codex-tty"
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