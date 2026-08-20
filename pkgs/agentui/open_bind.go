package agentui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

// OpenGrokBindState is durable bind progress for concurrent status probes.
// Stored at sessions/<session_id>/bind.json.
type OpenGrokBindState struct {
	State           string `json:"state"` // in_progress|ok|failed
	StartedAt       string `json:"started_at"`
	FinishedAt      string `json:"finished_at,omitempty"`
	Error           string `json:"error,omitempty"`
	RunnerSessionID string `json:"runner_session_id,omitempty"`
}

// openGrokBindResult is the outcome of the background open bind worker.
type openGrokBindResult struct {
	id          string
	updatesPath string
	err         error
	requireBind bool
}

// openGrokBindWorker runs DiscoverSession in the background during --open.
type openGrokBindWorker struct {
	done   <-chan openGrokBindResult
	cancel context.CancelFunc
}

// BindJSONPath returns the durable bind state path under the agent-run home.
func BindJSONPath(home, runner, sessionID string) string {
	return filepath.Join(home, "sessions", sessionID, "bind.json")
}

// ReadOpenGrokBindState loads bind.json if present.
func ReadOpenGrokBindState(home, runner, sessionID string) (OpenGrokBindState, bool) {
	data, err := os.ReadFile(BindJSONPath(home, runner, sessionID))
	if err != nil {
		return OpenGrokBindState{}, false
	}
	var st OpenGrokBindState
	if json.Unmarshal(data, &st) != nil {
		return OpenGrokBindState{}, false
	}
	st.State = strings.ToLower(strings.TrimSpace(st.State))
	return st, st.State != ""
}

func writeOpenGrokBindState(store agentstorage.Store, runner, sessionID string, st OpenGrokBindState) error {
	if store == nil {
		return fmt.Errorf("store is required")
	}
	dir := filepath.Join(store.Home(), "sessions", sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	path := BindJSONPath(store.Home(), runner, sessionID)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// openGrokBindPostDetachGrace is how long to keep discovering after attach
// returns when bind is still required and not yet done.
const openGrokBindPostDetachGrace = 20 * time.Second

// openGrokBindSoftTimeout is the full budget for empty-prompt / soft open only.
const openGrokBindSoftTimeout = 750 * time.Millisecond

// detachGrokBindTimeout is the product soft discovery budget for --detach
// when a real grok provider session is expected (miss → exit 0).
const detachGrokBindTimeout = 1 * time.Minute

// detachGrokBindSoftMissTimeout is a short soft budget when discovery is not
// expected to succeed (fake TUI override without GROK_HOME / empty prompt).
const detachGrokBindSoftMissTimeout = openGrokBindSoftTimeout

// startOpenGrokBindWorker begins background discovery and marks bind state
// in_progress immediately so concurrent status can report "binding".
//
// Hard-require workers keep discovering for the whole --open lifetime (no short
// absolute deadline from start — user may stay attached for minutes). Soft
// workers use a short timeout. Callers always Wait() after attach returns.
func startOpenGrokBindWorker(opts RunOptions, runner, sessionID, workspace, prompt string, runStart time.Time, knownID string) *openGrokBindWorker {
	requireBind := openGrokDiscoveryRequired(opts)
	var ctx context.Context
	var cancel context.CancelFunc
	if requireBind {
		// Cancel-only: discovery continues until success, Cancel(), or Wait grace.
		ctx, cancel = context.WithCancel(context.Background())
	} else {
		// Empty-prompt open: do not block detach for long discovery.
		ctx, cancel = context.WithTimeout(context.Background(), openGrokBindSoftTimeout)
	}

	done := make(chan openGrokBindResult, 1)
	startedAt := time.Now().UTC().Format(time.RFC3339Nano)
	_ = writeOpenGrokBindState(opts.Store, runner, sessionID, OpenGrokBindState{
		State:     "in_progress",
		StartedAt: startedAt,
	})

	go func() {
		defer cancel()
		res := runOpenGrokBind(ctx, opts, runner, sessionID, workspace, prompt, runStart, knownID, requireBind, startedAt)
		done <- res
	}()
	return &openGrokBindWorker{done: done, cancel: cancel}
}

// detachGrokBindBudget returns how long --detach should soft-wait for grok discovery.
// Product max is detachGrokBindTimeout when a real provider session is expected;
// fake-TUI / empty-prompt paths soft-miss quickly (still exit 0).
func detachGrokBindBudget(opts RunOptions) time.Duration {
	if detachUsesFullGrokBindBudget(opts) {
		return detachGrokBindTimeout
	}
	// Empty prompt + no hook: skip meaningful wait (reopen-only).
	if strings.TrimSpace(opts.Prompt) == "" && strings.TrimSpace(os.Getenv("AGENT_RUN_GROK_TTY_GROK_SESSION_ID")) == "" {
		return 0
	}
	return detachGrokBindSoftMissTimeout
}

// detachUsesFullGrokBindBudget reports whether detach soft-bind should use the
// full 1-minute product budget (real grok is expected to write sessions).
func detachUsesFullGrokBindBudget(opts RunOptions) bool {
	if strings.TrimSpace(opts.AgentRunnerConfigHome) != "" {
		return true
	}
	if strings.TrimSpace(os.Getenv("GROK_HOME")) != "" {
		return true
	}
	if strings.TrimSpace(os.Getenv("AGENT_RUNNER_CONFIG_HOME")) != "" {
		return true
	}
	if strings.TrimSpace(os.Getenv("AGENT_RUN_GROK_TTY_GROK_SESSION_ID")) != "" {
		return true
	}
	// No fake TUI command override → real grok binary path; allow full soft budget.
	if strings.TrimSpace(os.Getenv("AGENT_RUN_GROK_TTY_COMMAND")) == "" {
		return true
	}
	return false
}

// startDetachGrokBindWorker begins soft-only discovery for --detach.
// Never hard-fails on miss (requireBind=false). Returns nil when budget is 0
// (empty-prompt skip). Miss still exit 0 at the caller.
func startDetachGrokBindWorker(opts RunOptions, runner, sessionID, workspace, prompt string, runStart time.Time, knownID string) *openGrokBindWorker {
	budget := detachGrokBindBudget(opts)
	if budget <= 0 {
		return nil
	}
	// Soft-only: miss always exits 0 for detach.
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	const requireBind = false

	done := make(chan openGrokBindResult, 1)
	startedAt := time.Now().UTC().Format(time.RFC3339Nano)
	_ = writeOpenGrokBindState(opts.Store, runner, sessionID, OpenGrokBindState{
		State:     "in_progress",
		StartedAt: startedAt,
	})

	go func() {
		defer cancel()
		res := runOpenGrokBind(ctx, opts, runner, sessionID, workspace, prompt, runStart, knownID, requireBind, startedAt)
		// Detach never surfaces hard bind errors.
		res.err = nil
		res.requireBind = false
		done <- res
	}()
	return &openGrokBindWorker{done: done, cancel: cancel}
}

// Wait joins the worker. If discovery is still running (hard path during a long
// attach), keeps waiting up to openGrokBindPostDetachGrace after this call,
// then cancels and returns the final result.
func (w *openGrokBindWorker) Wait() openGrokBindResult {
	if w == nil {
		return openGrokBindResult{}
	}
	select {
	case res := <-w.done:
		return res
	case <-time.After(openGrokBindPostDetachGrace):
		w.Cancel()
		return <-w.done
	}
}

// WaitDone joins the worker until it finishes (ctx timeout/cancel already set).
// Used by --detach so the full soft budget can elapse without the open grace cancel.
func (w *openGrokBindWorker) WaitDone() openGrokBindResult {
	if w == nil {
		return openGrokBindResult{}
	}
	return <-w.done
}

func (w *openGrokBindWorker) Cancel() {
	if w != nil && w.cancel != nil {
		w.cancel()
	}
}

func runOpenGrokBind(ctx context.Context, opts RunOptions, runner, sessionID, workspace, prompt string, runStart time.Time, knownID string, requireBind bool, startedAt string) openGrokBindResult {
	res := openGrokBindResult{requireBind: requireBind}
	grokHome := agenttty.GrokHomeForRunner(opts.AgentRunnerConfigHome)
	id := strings.TrimSpace(knownID)
	updatesPath := ""
	debugOpenBind("bind start session=%s grokHome=%q workspace=%q knownID=%q require=%v GROK_HOME=%q HOME=%q configHome=%q",
		sessionID, grokHome, workspace, id, requireBind,
		os.Getenv("GROK_HOME"), os.Getenv("HOME"), opts.AgentRunnerConfigHome)
	if id != "" {
		if path, ok := agenttty.FindUpdatesBySessionID(grokHome, workspace, id); ok {
			updatesPath = path
		}
	}
	// Unbound: discover provider session. Already bound (resume with
	// meta.runner_session_id): accept knownID even when updates.jsonl is absent
	// under this GROK_HOME (fake TUI / reclaim fixtures) so hard-require open
	// does not fail after zombie terminal reclaim.
	if id == "" {
		discoverPrompt := strings.TrimSpace(opts.Prompt)
		if discoverPrompt == "" {
			discoverPrompt = strings.TrimSpace(prompt)
		}
		discID, path, discErr := agenttty.DiscoverSession(ctx, grokHome, workspace, discoverPrompt, runStart)
		if discErr != nil || strings.TrimSpace(discID) == "" {
			msg := "grok session id not resolved"
			if discErr != nil {
				msg = discErr.Error()
			}
			// Snapshot discovery inputs for parallel-flake diagnosis (always on
			// failure; also gated debug to stderr when AGENT_RUN_DEBUG_OPEN_BIND=1).
			sessionsRoot := filepath.Join(grokHome, "sessions")
			entries, _ := os.ReadDir(sessionsRoot)
			var entryNames []string
			for i, e := range entries {
				if i >= 8 {
					entryNames = append(entryNames, "…")
					break
				}
				entryNames = append(entryNames, e.Name())
			}
			debugOpenBind("bind FAIL session=%s grokHome=%q workspace=%q prompt=%q runStart=%s err=%v sessions=%v",
				sessionID, grokHome, workspace, discoverPrompt, runStart.UTC().Format(time.RFC3339Nano), discErr, entryNames)
			finished := time.Now().UTC().Format(time.RFC3339Nano)
			_ = writeOpenGrokBindState(opts.Store, runner, sessionID, OpenGrokBindState{
				State:      "failed",
				StartedAt:  startedAt,
				FinishedAt: finished,
				Error:      msg,
			})
			// Hard-wait may still soft-unbound when only non-empty prompt forced
			// discovery (attach-first open without explicit GROK_HOME / hook).
			if openGrokHardFailOnUnresolved(opts) {
				res.err = fmt.Errorf("error: grok session id not resolved for session %s (grokHome=%s discErr=%v sessionsRootEntries=%d)",
					sessionID, grokHome, discErr, len(entries))
				_ = agentstorage.AppendErrorLog(opts.Store.Home(), sessionID, "grok-session-discovery", res.err)
			}
			return res
		}
		id = strings.TrimSpace(discID)
		updatesPath = path
		debugOpenBind("bind OK session=%s id=%s path=%s", sessionID, id, updatesPath)
	}
	if updatesPath != "" {
		if abs, absErr := filepath.Abs(updatesPath); absErr == nil {
			updatesPath = abs
		}
	}
	// Persist runner id ASAP for mid-open status / concurrent probes.
	_ = opts.Store.UpdateSessionRunnerSessionID(sessionID, id)
	finished := time.Now().UTC().Format(time.RFC3339Nano)
	_ = writeOpenGrokBindState(opts.Store, runner, sessionID, OpenGrokBindState{
		State:           "ok",
		StartedAt:       startedAt,
		FinishedAt:      finished,
		RunnerSessionID: id,
	})
	res.id = id
	res.updatesPath = updatesPath
	return res
}

// debugOpenBind writes opt-in diagnostic traces to a temp file rather than an
// attached provider TTY when AGENT_RUN_DEBUG_OPEN_BIND is set.
func debugOpenBind(format string, args ...any) {
	if strings.TrimSpace(os.Getenv("AGENT_RUN_DEBUG_OPEN_BIND")) == "" {
		return
	}
	msg := fmt.Sprintf(format, args...)
	f, err := os.OpenFile(filepath.Join(os.TempDir(), "agent-run-open-bind-debug.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s agent-run open-bind: %s\n", time.Now().UTC().Format(time.RFC3339Nano), msg)
}

// printOpenGrokBindResult returns a hard error when require-bind discovery
// failed. Successful bind state is durable in meta.json and bind.json; it is
// deliberately not printed into the terminal that hosted the provider TTY.
func printOpenGrokBindResult(res openGrokBindResult, stderr io.Writer) error {
	if res.err != nil {
		return res.err
	}
	id := strings.TrimSpace(res.id)
	if id == "" {
		// Soft unbound: no session lines.
		return nil
	}
	return nil
}
