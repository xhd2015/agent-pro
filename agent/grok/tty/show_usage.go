package tty

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	grokagent "github.com/xhd2015/agent-pro/agent/cli/grok"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/debuglog"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/creack/pty"
	"github.com/hinshun/vt10x"
)

const (
	envShowUsageCommand = "GROK_SHOW_USAGE_COMMAND"
	envShowUsageTimeout = "GROK_SHOW_USAGE_TIMEOUT"
	envShowUsageDebug   = "GROK_SHOW_USAGE_DEBUG"
	envShowUsageRetries = "GROK_SHOW_USAGE_RETRIES"

	defaultTimeoutSeconds = 60
	defaultMaxAttempts    = 0 // 0 = retry until context deadline
	idleDebounce          = 200 * time.Millisecond
	pollInterval          = 50 * time.Millisecond
	promptSettleDelay     = 100 * time.Millisecond
	promptStableDelay     = 300 * time.Millisecond
	retryBackoff          = 500 * time.Millisecond
	ptyRows               = 40
	ptyCols               = 120
)

// UsageInfo holds parsed grok /usage show output.
type UsageInfo struct {
	WeeklyLimit string
	NextReset   string
}

var (
	// Legacy: "Weekly limit: 65%"
	weeklyLimitLegacyRe = regexp.MustCompile(`(?i)weekly\s*limit:\s*(\d+%)`)
	// Panel header (Grok 1.0.3): "Weekly limit (SuperGrok Heavy)" — percent on following lines.
	weeklyLimitHeaderRe = regexp.MustCompile(`(?i)weekly\s*limit\b`)
	percentTokenRe      = regexp.MustCompile(`(\d+%)`)
	// nextResetCandidates: ordered formats; first match wins.
	// Whitelist known TZs only (PT, UTC) — no catch-all [A-Z]{2,4} (matches junk like "Imag").
	// Legacy "Next reset:" then panel "Resets:" / "Reset:".
	// No-timezone form is last per family; normalize keeps bare wall clock (local for consumers).
	nextResetCandidates = []*regexp.Regexp{
		regexp.MustCompile(`(?i)next\s*reset:\s*([A-Za-z]+\s*\d{1,2},\s*\d{1,2}:\d{2}\s*PT)`),
		regexp.MustCompile(`(?i)next\s*reset:\s*([A-Za-z]+\s*\d{1,2},\s*\d{1,2}:\d{2}\s*UTC)`),
		regexp.MustCompile(`(?i)next\s*reset:\s*([A-Za-z]+\s*\d{1,2},\s*\d{1,2}:\d{2})`),
		regexp.MustCompile(`(?i)resets?:\s*([A-Za-z]+\s*\d{1,2},\s*\d{1,2}:\d{2}\s*PT)`),
		regexp.MustCompile(`(?i)resets?:\s*([A-Za-z]+\s*\d{1,2},\s*\d{1,2}:\d{2}\s*UTC)`),
		regexp.MustCompile(`(?i)resets?:\s*([A-Za-z]+\s*\d{1,2},\s*\d{1,2}:\d{2})`),
	}
	// Matches Month Day, HH:MM with optional PT|UTC after spaces are stripped.
	resetDateRe = regexp.MustCompile(`(?i)^([A-Za-z]+)(\d{1,2}),(\d{1,2}:\d{2})(PT|UTC)?$`)
)

// Options configures FetchUsage.
// Command/TimeoutSeconds override env hooks so tests need not Setenv under Parallel.
type Options struct {
	Debug       bool
	MaxAttempts int
	// Command overrides GROK_SHOW_USAGE_COMMAND.
	Command string
	// TimeoutSeconds overrides GROK_SHOW_USAGE_TIMEOUT when > 0.
	TimeoutSeconds int
}

// FetchUsage launches grok in a PTY, submits /usage show, and parses the response.
func FetchUsage(ctx context.Context) (*UsageInfo, error) {
	return FetchUsageWithOptions(ctx, Options{
		Debug:       debugEnabled(),
		MaxAttempts: maxAttemptsFromEnv(),
	})
}

// FetchUsageWithOptions launches grok with optional verbose logging and retries.
func FetchUsageWithOptions(ctx context.Context, opts Options) (*UsageInfo, error) {
	v := newVerboseLog(opts.Debug)
	debuglog.Log(debuglog.Entry{
		Event: "fetch_start",
		Labels: map[string]string{
			"component": "grok_tty",
			"provider":  "grok",
			"phase":     "fetch",
		},
	})
	attempts := opts.MaxAttempts
	if attempts <= 0 {
		attempts = defaultMaxAttempts
	}

	env := newExecEnv()
	argv, err := buildArgv(env, opts)
	if err != nil {
		return nil, err
	}
	v.command(argv)

	var lastErr error
	for attempt := 1; attempts <= 0 || attempt <= attempts; attempt++ {
		if attempt > 1 {
			v.attempt(attempt)
		}
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return nil, fmt.Errorf("timeout waiting for usage output: %w", lastErr)
			}
			return nil, timeoutErr(ctx)
		default:
		}

		info, err := fetchUsageOnce(ctx, argv, v)
		if err == nil {
			v.stateChange("done", fmt.Sprintf("weekly=%s reset=%s", info.WeeklyLimit, info.NextReset))
			debuglog.Log(debuglog.Entry{
				Event: "fetch_done",
				Labels: map[string]string{
					"component": "grok_tty",
					"provider":  "grok",
					"phase":     "fetch",
				},
				Fields: map[string]any{
					"weekly_limit": info.WeeklyLimit,
					"next_reset":   info.NextReset,
					"attempt":      attempt,
				},
			})
			return info, nil
		}
		lastErr = err
		v.stateChange("attempt-failed", err.Error())
		debuglog.Log(debuglog.Entry{
			Event: "fetch_attempt_failed",
			Labels: map[string]string{
				"component": "grok_tty",
				"provider":  "grok",
				"phase":     "error",
			},
			Fields: map[string]any{
				"error":   err.Error(),
				"attempt": attempt,
			},
		})
		if isNonRetryable(err) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, lastErr
		default:
		}
		v.stateChange("retry", "restarting grok session")
		time.Sleep(retryBackoff)
	}
	return nil, lastErr
}

// ParseShowUsageOutput extracts weekly limit and next reset from grok scrollback or stdout.
func ParseShowUsageOutput(stdout string) (*UsageInfo, error) {
	return parseUsageText(stdout)
}

// FormatUsageLines returns the canonical two-line CLI output for usage info.
func FormatUsageLines(info *UsageInfo) string {
	if info == nil {
		return ""
	}
	return fmt.Sprintf("Weekly limit: %s\nNext reset: %s\n", info.WeeklyLimit, info.NextReset)
}

func fetchUsageOnce(ctx context.Context, argv []string, v *verboseLog) (*UsageInfo, error) {
	scrollback, err := capturePTY(ctx, argv, v)
	if err != nil {
		return nil, err
	}
	return parseUsage(scrollback)
}

func maxAttemptsFromEnv() int {
	v := strings.TrimSpace(os.Getenv(envShowUsageRetries))
	if v == "" {
		return defaultMaxAttempts // 0 = retry until context deadline
	}
	n, err := strconvAtoi(v)
	if err != nil || n <= 0 {
		return defaultMaxAttempts
	}
	return n
}

func strconvAtoi(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func newExecEnv() *agentexec.Env {
	return agentexec.NewEnv(&agentexec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")
}

func debugEnabled() bool {
	v := strings.TrimSpace(os.Getenv(envShowUsageDebug))
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}

func buildArgv(env *agentexec.Env, opts Options) ([]string, error) {
	if hook := strings.TrimSpace(opts.Command); hook != "" {
		return agenttty.ParseShellWords(hook)
	}
	if hook := strings.TrimSpace(os.Getenv(envShowUsageCommand)); hook != "" {
		return agenttty.ParseShellWords(hook)
	}

	path, err := registry.ResolveConfiguredCLIPath(
		"",
		registry.GrokCLIPathSettingKey,
		registry.EnvGrokCLIPath,
		"",
		func() (string, error) { return grokagent.FindAgentPath(env) },
	)
	if err != nil {
		return nil, fmt.Errorf("grok not found: %w", err)
	}
	return []string{path, "--trust", "--always-approve", "--permission-mode=bypassPermissions"}, nil
}

func argvHasTrust(argv []string) bool {
	for _, arg := range argv {
		if arg == "--trust" {
			return true
		}
	}
	return false
}

func capturePTY(ctx context.Context, argv []string, v *verboseLog) ([]byte, error) {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("start grok: %w", err)
	}
	_ = pty.Setsize(ptmx, &pty.Winsize{Rows: ptyRows, Cols: ptyCols})

	var scrollback bytes.Buffer
	var mu sync.Mutex
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		_, _ = io.Copy(&lockedWriter{mu: &mu, w: &scrollback}, ptmx)
	}()

	var waitOnce sync.Once
	var waitErr error
	waitCmd := func() error {
		waitOnce.Do(func() {
			waitErr = cmd.Wait()
		})
		return waitErr
	}

	defer shutdownPTY(ptmx, cmd, readDone, waitCmd)

	getScrollback := func() []byte {
		mu.Lock()
		defer mu.Unlock()
		return append([]byte(nil), scrollback.Bytes()...)
	}

	if argvHasTrust(argv) {
		v.stateChange("workspace-trusted", "via --trust flag")
	}
	v.stateChange("waiting-prompt", "launching grok TUI")
	if err := waitForPrompt(ctx, ptmx, getScrollback, v); err != nil {
		v.printf("prompt wait failed; scrollback:\n%s", formatScrollbackDebug(getScrollback()))
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, timeoutErr(ctx)
	case <-time.After(promptSettleDelay):
	}

	screen := currentScreen(getScrollback())
	if usageFieldsPresentOnScreen(screen) {
		v.stateChange("usage-on-screen", "skipping /usage show")
		return getScrollback(), nil
	}

	if _, err := ptmx.Write([]byte("/usage show\r")); err != nil {
		return nil, fmt.Errorf("write /usage show: %w", err)
	}
	v.stateChange("submitted", "/usage show")

	if err := waitForUsageResponse(ctx, getScrollback, v, screen, readDone, waitCmd); err != nil {
		v.printf("usage wait failed; scrollback:\n%s", formatScrollbackDebug(getScrollback()))
		return nil, err
	}
	return getScrollback(), nil
}

func shutdownPTY(ptmx *os.File, cmd *exec.Cmd, readDone <-chan struct{}, waitCmd func() error) {
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	_ = ptmx.Close()
	select {
	case <-readDone:
	case <-time.After(time.Second):
	}
	if cmd == nil || cmd.Process == nil {
		return
	}
	done := make(chan struct{})
	go func() {
		_ = waitCmd()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
}

func formatScrollbackDebug(scrollback []byte) string {
	plain := agenttty.StripANSI(scrollback)
	rendered := renderScrollback(scrollback)
	if rendered != "" {
		return "--- rendered ---\n" + rendered + "\n--- plain ---\n" + plain
	}
	return plain
}

func waitForPrompt(ctx context.Context, ptmx io.Writer, getScrollback func() []byte, v *verboseLog) error {
	var readySince time.Time
	lastStarting := false
	lastReady := false
	lastTrust := false

	for {
		select {
		case <-ctx.Done():
			return timeoutErr(ctx)
		default:
		}

		sb := getScrollback()
		trustVisible := trustPromptVisible(sb)
		if trustVisible != lastTrust {
			if trustVisible {
				v.stateChange("asking-trust-for-workspace", "")
				v.stateChange("accepting-trust", "pressing Enter")
				if _, err := ptmx.Write([]byte("\r")); err != nil {
					if postTrustReady(sb) {
						trustVisible = false
					} else {
						return fmt.Errorf("accept workspace trust: %w", err)
					}
				}
				readySince = time.Time{}
				lastReady = false
			} else if lastTrust {
				v.stateChange("workspace-trusted", "")
				v.stateChange("waiting-prompt", "after workspace trust")
			}
			lastTrust = trustVisible
		}
		if trustVisible {
			time.Sleep(pollInterval)
			continue
		}

		screen := currentScreen(sb)
		starting := sessionStartingOnScreen(screen)
		ready := grokReadyOnScreen(screen)

		if starting != lastStarting {
			if starting {
				v.stateChange("session-starting", "grok booting")
			} else if lastStarting {
				v.stateChange("session-ready", "startup complete")
			}
			lastStarting = starting
		}

		if ready != lastReady {
			if ready {
				v.stateChange("prompt-ready", "grok input prompt visible")
			} else if lastReady {
				v.stateChange("prompt-lost", "waiting for prompt")
			}
			lastReady = ready
		}

		if ready {
			if readySince.IsZero() {
				readySince = time.Now()
			} else if time.Since(readySince) >= promptStableDelay {
				v.stateChange("prompt-stable", fmt.Sprintf("prompt stable for %s", promptStableDelay))
				return nil
			}
		} else {
			readySince = time.Time{}
		}

		time.Sleep(pollInterval)
	}
}

func currentScreen(scrollback []byte) string {
	return renderScrollback(scrollback)
}

func trustPromptVisible(scrollback []byte) bool {
	screen := currentScreen(scrollback)
	plain := agenttty.StripANSI(scrollback)
	if !trustPromptInText(screen) && !trustPromptInText(plain) {
		return false
	}
	return !postTrustReady(scrollback)
}

func postTrustReady(scrollback []byte) bool {
	return promptPastTrustScreen(currentScreen(scrollback)) ||
		promptPastTrustScreen(agenttty.StripANSI(scrollback))
}

func promptPastTrustScreen(text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	if usageFieldsPresentOnScreen(text) {
		return true
	}
	if strings.Contains(text, "Grok ›") || strings.Contains(text, "Grok \u203a") {
		return true
	}
	lower := strings.ToLower(text)
	return strings.Contains(lower, "composer") && promptDetectedInText(text)
}

func trustPromptInText(text string) bool {
	compact := compactScreenText(text)
	if strings.Contains(compact, "doyoutrustthecontentsofthisdirectory") {
		return strings.Contains(compact, "yes,proceed") ||
			strings.Contains(compact, "yes,continue") ||
			strings.Contains(compact, "pressentertocontinue") ||
			strings.Contains(compact, "grokbuildmayrunormodify")
	}
	return strings.Contains(compact, "workspacetrust") && strings.Contains(compact, "yes,proceed")
}

func compactScreenText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range strings.ToLower(s) {
		switch r {
		case ' ', '\n', '\r', '\t', '\b', '\f':
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func sessionStartingOnScreen(screen string) bool {
	lower := strings.ToLower(screen)
	return strings.Contains(lower, "starting session") || strings.Contains(lower, "startingsession")
}

func grokReadyOnScreen(screen string) bool {
	if strings.TrimSpace(screen) == "" {
		return false
	}
	if trustPromptInText(screen) && !promptPastTrustScreen(screen) {
		return false
	}
	if usageFieldsPresentOnScreen(screen) {
		return true
	}
	if sessionStartingOnScreen(screen) {
		return false
	}
	if !promptDetectedInText(screen) {
		return false
	}
	lower := strings.ToLower(screen)
	// Real grok: main UI footer is visible when the session accepts input.
	if strings.Contains(lower, "composer") || strings.Contains(lower, "always-approve") {
		return true
	}
	// Fake TUI used by doctests.
	return strings.Contains(screen, "Grok ›") || strings.Contains(screen, "Grok \u203a")
}

func waitForUsageResponse(ctx context.Context, getScrollback func() []byte, v *verboseLog, baselineScreen string, readDone <-chan struct{}, waitCmd func() error) error {
	var parseableSince time.Time
	var screenIdleSince time.Time
	lastParseable := false
	lastScreen := ""

	for {
		select {
		case <-ctx.Done():
			return timeoutErr(ctx)
		case <-readDone:
			if err := waitCmd(); err != nil {
				return fmt.Errorf("grok process exited: %w", err)
			}
			if _, parseErr := parseUsage(getScrollback()); parseErr != nil {
				v.stateChange("parse-failed", parseErr.Error())
				return parseErr
			}
			v.stateChange("usage-ready", "usage fields present after process exit")
			return nil
		default:
		}

		screen := currentScreen(getScrollback())
		if screen != lastScreen {
			lastScreen = screen
			screenIdleSince = time.Now()
		}

		parseable := usageFieldsPresentOnScreen(screen)

		if parseable != lastParseable {
			if parseable {
				v.stateChange("usage-detected", "weekly limit and next reset on screen")
			}
			lastParseable = parseable
		}

		if parseable {
			if parseableSince.IsZero() {
				parseableSince = time.Now()
			} else if time.Since(parseableSince) >= idleDebounce {
				v.stateChange("usage-ready", "usage fields stable on screen")
				return nil
			}
		} else {
			parseableSince = time.Time{}
			if !screenIdleSince.IsZero() && time.Since(screenIdleSince) >= idleDebounce &&
				malformedUsageResponse(screen) {
				_, err := parseUsage(getScrollback())
				v.stateChange("parse-failed", err.Error())
				return err
			}
		}

		time.Sleep(pollInterval)
	}
}

func isNonRetryable(err error) bool {
	return err != nil && strings.Contains(err.Error(), "failed to parse usage output")
}

func malformedUsageResponse(screen string) bool {
	return strings.Contains(strings.ToLower(screen), "not usage")
}

func usageFieldsPresentOnScreen(screen string) bool {
	if strings.TrimSpace(screen) == "" {
		return false
	}
	_, err := parseUsageText(screen)
	return err == nil
}

func renderScrollback(scrollback []byte) string {
	if len(scrollback) == 0 {
		return ""
	}
	vt := vt10x.New(vt10x.WithSize(ptyCols, ptyRows))
	if _, err := vt.Write(scrollback); err != nil {
		return ""
	}
	vt.Lock()
	defer vt.Unlock()

	var out strings.Builder
	for y := 0; y < ptyRows; y++ {
		runes := make([]rune, ptyCols)
		lastNonSpace := -1
		for x := 0; x < ptyCols; x++ {
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
			continue
		}
		out.WriteString(string(runes[:lastNonSpace+1]))
		out.WriteByte('\n')
	}
	return out.String()
}

func promptDetectedInText(plain string) bool {
	if strings.Contains(plain, "Grok ›") || strings.Contains(plain, "Grok \u203a") {
		return true
	}
	lower := strings.ToLower(plain)
	if strings.Contains(lower, "grok build") && (strings.Contains(plain, "❯") || strings.Contains(plain, "\u276f") || strings.Contains(plain, "›")) {
		return true
	}
	return strings.Contains(plain, "❯") || strings.Contains(plain, "\u276f")
}

func parseUsage(scrollback []byte) (*UsageInfo, error) {
	return parseUsageText(usageCorpus(scrollback))
}

func parseUsageText(corpus string) (*UsageInfo, error) {
	weekly, weeklyOK := matchWeeklyLimit(corpus)
	nextRaw, nextOK := matchNextReset(corpus)
	if !weeklyOK || !nextOK {
		return nil, fmt.Errorf("failed to parse usage output")
	}
	return &UsageInfo{
		WeeklyLimit: weekly,
		NextReset:   normalizeResetDate(nextRaw),
	}, nil
}

// matchWeeklyLimit accepts legacy "Weekly limit: N%" or modal panel form:
// "Weekly limit (...)" header then first N% on a following line.
func matchWeeklyLimit(corpus string) (string, bool) {
	if m := weeklyLimitLegacyRe.FindStringSubmatch(corpus); len(m) >= 2 {
		return strings.TrimSpace(m[1]), true
	}
	loc := weeklyLimitHeaderRe.FindStringIndex(corpus)
	if loc == nil {
		return "", false
	}
	// Percent is on following lines (progress bar), not the header line itself.
	rest := corpus[loc[1]:]
	nl := strings.IndexByte(rest, '\n')
	if nl < 0 {
		return "", false
	}
	rest = rest[nl+1:]
	if m := percentTokenRe.FindStringSubmatch(rest); len(m) >= 2 {
		return m[1], true
	}
	return "", false
}

// matchNextReset tries ordered next-reset candidates; first match wins.
func matchNextReset(corpus string) (string, bool) {
	for _, re := range nextResetCandidates {
		if m := re.FindStringSubmatch(corpus); len(m) >= 2 {
			return strings.TrimSpace(m[1]), true
		}
	}
	return "", false
}

func usageCorpus(scrollback []byte) string {
	plain := agenttty.StripANSI(scrollback)
	rendered := renderScrollback(scrollback)
	if rendered != "" {
		return plain + "\n" + rendered
	}
	return plain
}

// normalizeResetDate reformats Month Day, HH:MM [TZ] canonically.
// Missing timezone is left bare (local wall clock for consumers; Grok 0.2.99 omits TZ).
func normalizeResetDate(raw string) string {
	raw = strings.TrimSpace(raw)
	compact := strings.ReplaceAll(raw, " ", "")
	m := resetDateRe.FindStringSubmatch(compact)
	if len(m) < 4 {
		return raw
	}
	if len(m) >= 5 && m[4] != "" {
		tz := strings.ToUpper(m[4])
		return fmt.Sprintf("%s %s, %s %s", m[1], m[2], m[3], tz)
	}
	// No timezone: bare local wall-clock time.
	return fmt.Sprintf("%s %s, %s", m[1], m[2], m[3])
}

func timeoutErr(ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("timeout waiting for usage output")
	}
	return fmt.Errorf("timeout waiting for usage output: %w", ctx.Err())
}

type lockedWriter struct {
	mu *sync.Mutex
	w  io.Writer
}

func (lw *lockedWriter) Write(p []byte) (int, error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	return lw.w.Write(p)
}