package tty

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	codexagent "github.com/xhd2015/agent-pro/agent/cli/codex"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/debuglog"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

const (
	envShowStatusCommand   = "CODEX_SHOW_STATUS_COMMAND"
	envShowStatusTimeout   = "CODEX_SHOW_STATUS_TIMEOUT"
	envShowStatusSessionID = "CODEX_SHOW_STATUS_SESSION_ID"
	envShowStatusDebug     = "CODEX_SHOW_STATUS_DEBUG"
	envTTYWatchHome        = "TTY_WATCH_HOME"

	defaultSessionID       = "codex-status-usage"
	defaultTimeoutSeconds  = 60
	pollInterval           = 1 * time.Second
	retryPollInterval      = 300 * time.Millisecond
)

// UsageInfo holds parsed Codex /status output.
type UsageInfo struct {
	MonthlyUsage string
	CreditsUsed  string
	CreditsTotal string
	NextReset    string
}

// Options configures FetchStatusWithOptions.
type Options struct {
	Debug bool
}

var (
	percentLeftRe = regexp.MustCompile(`(?i)monthly\s+credit\s+limit:.*?(\d+)%\s*left`)
	resetRe       = regexp.MustCompile(`(?i)\(resets\s+([^)]+)\)`)
	creditsRe     = regexp.MustCompile(`(?i)([\d,]+)\s+of\s+([\d,]+)\s+credits\s+used`)
)

// FetchStatus launches Codex in an ephemeral ttywatch session and parses /status output.
func FetchStatus(ctx context.Context) (*UsageInfo, error) {
	return FetchStatusWithOptions(ctx, Options{Debug: debugEnabled()})
}

// FetchStatusWithOptions launches Codex with optional verbose logging.
func FetchStatusWithOptions(ctx context.Context, opts Options) (*UsageInfo, error) {
	v := newVerboseLog(opts.Debug)
	debuglog.Log(debuglog.Entry{
		Event: "fetch_start",
		Labels: map[string]string{
			"component": "codex_tty",
			"provider":  "codex",
			"phase":     "fetch",
		},
	})

	codexArgv, err := buildCodexArgv(newExecEnv())
	if err != nil {
		logCodexError("build_argv", err, nil)
		return nil, err
	}
	extraPaths := commandExtraPaths(codexArgv)
	v.command(append([]string{"ttywatch", "run", "--session-id", sessionIDFromEnv()}, codexArgv...))

	sessionID := sessionIDFromEnv()
	home, err := ttyWatchHome()
	if err != nil {
		logCodexError("tty_watch_home", err, nil)
		return nil, err
	}

	release, err := ttywatch.ReserveCustomSessionID(ttywatch.DefaultRegistryConfig(home), sessionID)
	if err != nil {
		logCodexError("reserve_session", err, map[string]any{"session_id": sessionID})
		return nil, err
	}
	// Release the registry flock immediately after reservation so concurrent
	// tty-watch / reserve peers are not blocked for the whole StartInProcess + wait.
	// Claim/registry still protect the session id until session.Kill().
	release()

	session := ttywatch.NewEphemeralSession(home, sessionID, codexArgv)
	session.ExtraPaths = extraPaths
	defer func() { _ = session.Kill() }()

	debuglog.Log(debuglog.Entry{
		Event: "session_start",
		Labels: map[string]string{
			"component": "codex_tty",
			"provider":  "codex",
			"phase":     "session",
		},
		Fields: map[string]any{
			"session_id":  sessionID,
			"tty_home":    home,
			"argv":        codexArgv,
			"extra_paths": extraPaths,
		},
	})

	v.stateChange("starting-session", strings.Join(codexArgv, " "))
	startPhase := time.Now()
	if err := session.StartInProcess(ctx); err != nil {
		logCodexError("start_session", err, map[string]any{"session_id": sessionID})
		return nil, fmt.Errorf("start ttywatch session: %w", err)
	}
	v.phaseDone("start-session", startPhase)
	v.stateChange("session-registered", sessionID)

	fetchDeadline := deadlineForFetch(ctx)

	waitPromptPhase := time.Now()
	if err := waitForPrompt(ctx, session, fetchDeadline, v); err != nil {
		logCodexError("wait_prompt", err, snapshotExcerptFields(session))
		return nil, err
	}
	v.phaseDone("wait-prompt", waitPromptPhase)

	if err := session.Send("/status\n\r"); err != nil {
		logCodexError("send_status", err, nil)
		return nil, fmt.Errorf("send /status: %w", err)
	}
	v.stateChange("submitted", "/status")
	debuglog.Log(debuglog.Entry{
		Event: "send_status",
		Labels: map[string]string{
			"component": "codex_tty",
			"provider":  "codex",
			"phase":     "status",
		},
	})

	waitStatusPhase := time.Now()
	snapshot, err := waitForStatusSnapshot(ctx, session, fetchDeadline, v)
	if err != nil {
		logCodexError("wait_status", err, snapshotExcerptFields(session))
		return nil, err
	}
	v.phaseDone("wait-status", waitStatusPhase)

	info, err := ParseStatusSnapshot(snapshot)
	if err != nil {
		logCodexError("parse_status", err, map[string]any{"snapshot_excerpt": excerptText(snapshot, 500)})
		return nil, err
	}
	v.stateChange("done", fmt.Sprintf("monthly=%s credits=%s/%s reset=%s",
		info.MonthlyUsage, info.CreditsUsed, info.CreditsTotal, info.NextReset))
	debuglog.Log(debuglog.Entry{
		Event: "fetch_done",
		Labels: map[string]string{
			"component": "codex_tty",
			"provider":  "codex",
			"phase":     "fetch",
		},
		Fields: map[string]any{
			"monthly_usage": info.MonthlyUsage,
			"credits_used":  info.CreditsUsed,
			"credits_total": info.CreditsTotal,
			"next_reset":    info.NextReset,
		},
	})
	return info, nil
}

// ParseStatusSnapshot extracts monthly usage, credits, and reset from tty-watch snapshot text.
func ParseStatusSnapshot(stdout string) (*UsageInfo, error) {
	return parseStatusText(stdout)
}

// FormatStatusLines returns the canonical three-line CLI output for status info.
func FormatStatusLines(info *UsageInfo) string {
	if info == nil {
		return ""
	}
	return fmt.Sprintf("Monthly usage: %s\nCredits used: %s of %s\nNext reset: %s\n",
		info.MonthlyUsage, info.CreditsUsed, info.CreditsTotal, info.NextReset)
}

func parseStatusText(corpus string) (*UsageInfo, error) {
	percent := percentLeftRe.FindStringSubmatch(corpus)
	reset := resetRe.FindStringSubmatch(corpus)
	credits := creditsRe.FindStringSubmatch(corpus)
	if len(percent) < 2 || len(reset) < 2 || len(credits) < 3 {
		return nil, fmt.Errorf("failed to parse status output")
	}

	left, err := strconv.Atoi(strings.TrimSpace(percent[1]))
	if err != nil {
		return nil, fmt.Errorf("failed to parse status output")
	}
	usage := 100 - left
	if usage < 0 {
		return nil, fmt.Errorf("failed to parse status output")
	}

	return &UsageInfo{
		MonthlyUsage: fmt.Sprintf("%d%%", usage),
		CreditsUsed:  stripCommas(strings.TrimSpace(credits[1])),
		CreditsTotal: stripCommas(strings.TrimSpace(credits[2])),
		NextReset:    strings.TrimSpace(reset[1]),
	}, nil
}

func stripCommas(s string) string {
	return strings.ReplaceAll(s, ",", "")
}

func newExecEnv() *agentexec.Env {
	return agentexec.NewEnv(&agentexec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")
}

func debugEnabled() bool {
	v := strings.TrimSpace(os.Getenv(envShowStatusDebug))
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}

func sessionIDFromEnv() string {
	if v := strings.TrimSpace(os.Getenv(envShowStatusSessionID)); v != "" {
		return v
	}
	return defaultSessionID
}

func ttyWatchHome() (string, error) {
	if v := os.Getenv(envTTYWatchHome); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tty-watch"), nil
}

// commandExtraPaths returns PATH entries the PTY child needs beyond the daemon PATH.
// npm-installed codex is a #!/usr/bin/env node shim; node must be on PATH at spawn time.
func commandExtraPaths(argv []string) []string {
	if len(argv) == 0 {
		return nil
	}
	bin := strings.TrimSpace(argv[0])
	if bin == "" {
		return nil
	}
	dir := filepath.Dir(bin)
	if dir == "" || dir == "." {
		return nil
	}
	return []string{dir}
}

func buildCodexArgv(env *agentexec.Env) ([]string, error) {
	if hook := strings.TrimSpace(os.Getenv(envShowStatusCommand)); hook != "" {
		return agenttty.ParseShellWords(hook)
	}

	path, err := registry.ResolveConfiguredCLIPath(
		"",
		registry.CodexCLIPathSettingKey,
		registry.EnvCodexCLIPath,
		"",
		func() (string, error) { return codexagent.FindAgentPath(env) },
	)
	if err != nil {
		return nil, fmt.Errorf("codex not found: %w", err)
	}
	// Ephemeral /status fetch does not need user-configured MCP servers; skipping
	// them avoids slow startup and failures from optional servers (e.g. computer-use).
	return []string{path, "--dangerously-bypass-approvals-and-sandbox", "-c", "mcp_servers={}"}, nil
}

func deadlineForFetch(ctx context.Context) time.Time {
	if ctxDeadline, ok := ctx.Deadline(); ok {
		return ctxDeadline
	}
	return time.Now().Add(timeoutFromEnv())
}

// Signed PROTOCOL.md keys for Codex Update available menu auto-Skip.
const (
	csiCursorDown = "\x1b[B" // hex 1b5b42
	keyEnterCR    = "\r"     // hex 0d
)

func waitForPrompt(ctx context.Context, session *ttywatch.EphemeralSession, deadline time.Time, v *verboseLog) error {
	provider, ok := agenttty.Get("codex-tty")
	if !ok {
		return fmt.Errorf("codex-tty provider not registered")
	}

	v.stateChange("waiting-prompt", fmt.Sprintf("codex prompt on screen (deadline=%s)", deadline.Format(time.RFC3339)))

	waitStart := time.Now()
	var lastPollLog time.Time
	pollCount := 0
	var snapshotTotal time.Duration
	updateSkipDone := false
	for {
		select {
		case <-ctx.Done():
			return timeoutErr(ctx)
		default:
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for status output")
		}

		pollCount++
		snapStart := time.Now()
		snapshot, err := session.Snapshot()
		snapDur := time.Since(snapStart)
		snapshotTotal += snapDur
		if err != nil {
			v.snapshot("wait-prompt", snapDur, "error: "+err.Error())
			if isRetryableSnapshotError(err) && time.Now().Before(deadline) {
				debuglog.Log(debuglog.Entry{
					Event: "snapshot_retry",
					Labels: map[string]string{
						"component": "codex_tty",
						"provider":  "codex",
						"phase":     "prompt",
					},
					Fields: map[string]any{
						"error":      err.Error(),
						"poll":       pollCount,
						"elapsed_ms": time.Since(waitStart).Milliseconds(),
					},
				})
				sleepUntilPoll(ctx, deadline, retryPollInterval)
				continue
			}
			return err
		}
		if statusFieldsPresent(snapshot) {
			v.snapshot("wait-prompt", snapDur, fmt.Sprintf("poll=%d status-already-visible", pollCount))
			v.stateChange("prompt-ready", fmt.Sprintf("status fields already visible (polls=%d snapshot_total=%s)", pollCount, snapshotTotal.Round(time.Millisecond)))
			return nil
		}

		// Auto-Skip the blocking Update available menu before treating the screen
		// as a normal wait-for-idle (signed PROTOCOL.md).
		if !updateSkipDone && agenttty.IsBlockingUpdateMenu(snapshot) {
			v.stateChange("update-menu", "blocking update menu detected; auto-Skip")
			if err := dismissBlockingUpdateMenu(ctx, session, deadline, v); err != nil {
				logCodexError("update_menu_skip", err, map[string]any{
					"snapshot_excerpt": excerptText(snapshot, 300),
				})
				return err
			}
			updateSkipDone = true
			// Resume polling for idle prompt after menu dismiss (residual banner OK).
			sleepUntilPoll(ctx, deadline, retryPollInterval)
			continue
		}

		st := provider.CheckWritable([]byte(snapshot))
		if st.Ready && st.State == "idle" {
			v.snapshot("wait-prompt", snapDur, fmt.Sprintf("poll=%d idle", pollCount))
			v.stateChange("prompt-ready", fmt.Sprintf("%s (polls=%d snapshot_total=%s)", st.State, pollCount, snapshotTotal.Round(time.Millisecond)))
			debuglog.Log(debuglog.Entry{
				Event: "prompt_ready",
				Labels: map[string]string{
					"component": "codex_tty",
					"provider":  "codex",
					"phase":     "prompt",
				},
				Fields: map[string]any{
					"writable":       st.Ready,
					"writable_state": st.State,
					"elapsed_ms":     time.Since(waitStart).Milliseconds(),
				},
			})
			return nil
		}

		if lastPollLog.IsZero() || time.Since(lastPollLog) >= 2*time.Second {
			lastPollLog = time.Now()
			v.snapshot("wait-prompt", snapDur, fmt.Sprintf("poll=%d state=%s ready=%v", pollCount, st.State, st.Ready))
			debuglog.Log(debuglog.Entry{
				Event: "wait_prompt",
				Labels: map[string]string{
					"component": "codex_tty",
					"provider":  "codex",
					"phase":     "prompt",
				},
				Fields: map[string]any{
					"writable":         st.Ready,
					"writable_state":   st.State,
					"writable_reason":  st.Reason,
					"elapsed_ms":       time.Since(waitStart).Milliseconds(),
					"snapshot_excerpt": excerptText(snapshot, 300),
				},
			})
		}

		sleepUntilPoll(ctx, deadline, pollInterval)
	}
}

// dismissBlockingUpdateMenu runs the signed Skip protocol:
// CSI Down → verify selection is SKIP (retry Down at most once) → Enter → poll until menu gone.
// Never sends Enter while selection is UPDATE_NOW.
func dismissBlockingUpdateMenu(ctx context.Context, session *ttywatch.EphemeralSession, deadline time.Time, v *verboseLog) error {
	// Initial Down (PROTOCOL step select_skip).
	if err := session.Send(csiCursorDown); err != nil {
		return fmt.Errorf("send CSI Down for update menu: %w", err)
	}
	v.stateChange("update-menu-down", "CSI Down (1)")

	// Poll briefly for selection to move to Skip (TUI re-render lag).
	sel, snapshot, err := waitForUpdateMenuSelection(ctx, session, deadline, v, "SKIP", 800*time.Millisecond)
	if err != nil {
		return err
	}
	if sel != "SKIP" {
		// Retry Down at most once (PROTOCOL: if assert fails after first Down).
		if err := session.Send(csiCursorDown); err != nil {
			// Prefer a clear Skip-selection error when inject is already flaky;
			// still never Enter while on Update now.
			return fmt.Errorf("could not select Skip on update prompt (selection=%q): %w", sel, err)
		}
		v.stateChange("update-menu-down", "CSI Down (retry)")
		sel, snapshot, err = waitForUpdateMenuSelection(ctx, session, deadline, v, "SKIP", 800*time.Millisecond)
		if err != nil {
			return err
		}
	}

	if sel != "SKIP" {
		// Never confirm Update now; fail early with a clear error.
		return fmt.Errorf("could not select Skip on update prompt (selection=%q) excerpt=%q",
			sel, excerptText(snapshot, 200))
	}

	// Enter only after verified Skip.
	if err := session.Send(keyEnterCR); err != nil {
		return fmt.Errorf("send Enter for update menu Skip: %w", err)
	}
	v.stateChange("update-menu-enter", "Enter on Skip")

	// Poll until menu options are gone (immediate frame may still show the menu).
	for {
		select {
		case <-ctx.Done():
			return timeoutErr(ctx)
		default:
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for update menu to dismiss")
		}
		snapshot, err = snapshotWithRetry(ctx, session, deadline)
		if err != nil {
			return err
		}
		if !agenttty.IsBlockingUpdateMenu(snapshot) {
			v.stateChange("update-menu-gone", "blocking update menu dismissed")
			debuglog.Log(debuglog.Entry{
				Event: "update_menu_dismissed",
				Labels: map[string]string{
					"component": "codex_tty",
					"provider":  "codex",
					"phase":     "prompt",
				},
			})
			return nil
		}
		sleepUntilPoll(ctx, deadline, retryPollInterval)
	}
}

// waitForUpdateMenuSelection polls snapshots until selection matches want or the
// wait window elapses. Returns the last selection and snapshot.
func waitForUpdateMenuSelection(ctx context.Context, session *ttywatch.EphemeralSession, deadline time.Time, v *verboseLog, want string, maxWait time.Duration) (string, string, error) {
	waitDeadline := time.Now().Add(maxWait)
	if waitDeadline.After(deadline) {
		waitDeadline = deadline
	}
	var lastSel string
	var lastSnap string
	for {
		select {
		case <-ctx.Done():
			return lastSel, lastSnap, timeoutErr(ctx)
		default:
		}
		if time.Now().After(deadline) {
			return lastSel, lastSnap, fmt.Errorf("timeout waiting for status output")
		}
		snapshot, err := snapshotWithRetry(ctx, session, deadline)
		if err != nil {
			return lastSel, lastSnap, err
		}
		lastSnap = snapshot
		lastSel = agenttty.UpdateMenuSelection(snapshot)
		if lastSel == want {
			return lastSel, lastSnap, nil
		}
		// Menu already gone (unexpected during select_skip) — stop waiting.
		if !agenttty.IsBlockingUpdateMenu(snapshot) {
			return lastSel, lastSnap, nil
		}
		if !time.Now().Before(waitDeadline) {
			return lastSel, lastSnap, nil
		}
		v.stateChange("update-menu-wait", fmt.Sprintf("selection=%q want=%q", lastSel, want))
		sleepUntilPoll(ctx, deadline, retryPollInterval)
	}
}

func snapshotWithRetry(ctx context.Context, session *ttywatch.EphemeralSession, deadline time.Time) (string, error) {
	for {
		select {
		case <-ctx.Done():
			return "", timeoutErr(ctx)
		default:
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("timeout waiting for status output")
		}
		snapshot, err := session.Snapshot()
		if err != nil {
			if isRetryableSnapshotError(err) && time.Now().Before(deadline) {
				sleepUntilPoll(ctx, deadline, retryPollInterval)
				continue
			}
			return "", err
		}
		return snapshot, nil
	}
}

func sleepUntilPoll(ctx context.Context, deadline time.Time, interval time.Duration) {
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return
	}
	if interval > remaining {
		interval = remaining
	}
	select {
	case <-ctx.Done():
	case <-time.After(interval):
	}
}

func timeoutFromEnv() time.Duration {
	timeout := defaultTimeoutSeconds
	if v := strings.TrimSpace(os.Getenv(envShowStatusTimeout)); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			timeout = sec
		}
	}
	return time.Duration(timeout) * time.Second
}

func codexPromptVisible(text string) bool {
	if strings.Contains(text, "Codex ›") || strings.Contains(text, "Codex \u203a") {
		return true
	}
	lower := strings.ToLower(text)
	if strings.Contains(lower, "openai codex") || strings.Contains(lower, "yolo mode") {
		return strings.Contains(text, "\u203a") || strings.Contains(text, "›")
	}
	return false
}

func statusFieldsPresent(text string) bool {
	_, err := parseStatusText(text)
	return err == nil
}

func waitForStatusSnapshot(ctx context.Context, session *ttywatch.EphemeralSession, deadline time.Time, v *verboseLog) (string, error) {
	v.stateChange("waiting-status", fmt.Sprintf("poll for /status output (deadline=%s)", deadline.Format(time.RFC3339)))
	var lastPollLog time.Time
	lastParseable := false
	pollCount := 0
	var snapshotTotal time.Duration

	for {
		select {
		case <-ctx.Done():
			return "", timeoutErr(ctx)
		default:
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("timeout waiting for status output")
		}

		pollCount++
		snapStart := time.Now()
		snapshot, err := session.Snapshot()
		snapDur := time.Since(snapStart)
		snapshotTotal += snapDur
		if err != nil {
			v.snapshot("wait-status", snapDur, "error: "+err.Error())
			if isRetryableSnapshotError(err) && time.Now().Before(deadline) {
				debuglog.Log(debuglog.Entry{
					Event: "snapshot_retry",
					Labels: map[string]string{
						"component": "codex_tty",
						"provider":  "codex",
						"phase":     "status",
					},
					Fields: map[string]any{
						"error": err.Error(),
						"poll":  pollCount,
					},
				})
				sleepUntilPoll(ctx, deadline, retryPollInterval)
				continue
			}
			return "", err
		}

		parseable := statusFieldsPresent(snapshot)
		if parseable != lastParseable {
			if parseable {
				v.snapshot("wait-status", snapDur, fmt.Sprintf("poll=%d parseable", pollCount))
				v.stateChange("status-detected", "monthly limit and credits on screen")
			}
			lastParseable = parseable
		}

		if parseable {
			v.stateChange("status-ready", fmt.Sprintf("polls=%d snapshot_total=%s", pollCount, snapshotTotal.Round(time.Millisecond)))
			debuglog.Log(debuglog.Entry{
				Event: "status_ready",
				Labels: map[string]string{
					"component": "codex_tty",
					"provider":  "codex",
					"phase":     "status",
				},
				Fields: map[string]any{
					"snapshot_excerpt": excerptText(snapshot, 300),
				},
			})
			return snapshot, nil
		}

		if malformedStatusResponse(snapshot) {
			_, err := parseStatusText(snapshot)
			v.stateChange("parse-failed", err.Error())
			return "", err
		}

		if lastPollLog.IsZero() || time.Since(lastPollLog) >= 2*time.Second {
			lastPollLog = time.Now()
			v.snapshot("wait-status", snapDur, fmt.Sprintf("poll=%d parseable=%v", pollCount, parseable))
			debuglog.Log(debuglog.Entry{
				Event: "wait_status",
				Labels: map[string]string{
					"component": "codex_tty",
					"provider":  "codex",
					"phase":     "status",
				},
				Fields: map[string]any{
					"parseable":        parseable,
					"snapshot_excerpt": excerptText(snapshot, 300),
				},
			})
		}

		sleepUntilPoll(ctx, deadline, pollInterval)
	}
}

func malformedStatusResponse(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "not status")
}

func isRetryableSnapshotError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "timeout waiting for snapshot frame") ||
		strings.Contains(msg, "timeout waiting for snapshot scrollback")
}

func timeoutErr(ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("timeout waiting for status output")
	}
	return fmt.Errorf("timeout waiting for status output: %w", ctx.Err())
}

func logCodexError(event string, err error, fields map[string]any) {
	if err == nil {
		return
	}
	if fields == nil {
		fields = map[string]any{}
	}
	fields["error"] = err.Error()
	debuglog.Log(debuglog.Entry{
		Event: event,
		Labels: map[string]string{
			"component": "codex_tty",
			"provider":  "codex",
			"phase":     "error",
		},
		Fields: fields,
	})
}

func snapshotExcerptFields(session *ttywatch.EphemeralSession) map[string]any {
	if session == nil {
		return nil
	}
	snapshot, err := session.Snapshot()
	if err != nil {
		return map[string]any{"snapshot_error": err.Error()}
	}
	return map[string]any{"snapshot_excerpt": excerptText(snapshot, 500)}
}

func excerptText(text string, max int) string {
	text = strings.ReplaceAll(text, "\r\n", "\\n")
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "\t", " ")
	if max > 0 && len(text) > max {
		return text[:max]
	}
	return text
}