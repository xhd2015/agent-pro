// Package agentrunapi provides in-process Classify + AutoSendOrResume for
// agent-run session lifecycle (run | send | resume) without shelling out to the
// agent-run binary.
package agentrunapi

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

// Mode is the auto-send-or-resume branch: run | send | resume.
type Mode string

const (
	ModeRun    Mode = "run"
	ModeSend   Mode = "send"
	ModeResume Mode = "resume"
)

// ProbeReport is the lifecycle subset Classify needs (from probeSessionStatus parity).
type ProbeReport struct {
	// ResumeReady is true when runner is bound and exited (resume path).
	ResumeReady bool
	// RunnerExited is nil when unknown; false when live; true when exited.
	RunnerExited *bool
}

// ProbeFunc injects lifecycle probing.
// nil → LifecycleProbe (production TTY/registry/process probe).
// Use EmptyProbe in unit tests that need unknown lifecycle without TTY I/O.
type ProbeFunc func(store agentstorage.Store, meta agentstorage.SessionMeta) (ProbeReport, error)

// Classify resolves session id and returns Mode using the same rules as
// cmd/agent-run runAutoSendOrResume. Missing session → ModeRun, found=false.
func Classify(store agentstorage.Store, sessionID string, probe ProbeFunc) (mode Mode, meta agentstorage.SessionMeta, found bool, err error) {
	if store == nil {
		return ModeRun, agentstorage.SessionMeta{}, false, fmt.Errorf("store is required")
	}
	sessionID = strings.TrimSpace(sessionID)
	meta, found, err = resolveSession(store, sessionID)
	if err != nil {
		return ModeRun, agentstorage.SessionMeta{}, false, err
	}
	if !found {
		return ModeRun, agentstorage.SessionMeta{}, false, nil
	}
	if probe == nil {
		probe = LifecycleProbe
	}
	report, err := probe(store, meta)
	if err != nil {
		return ModeRun, meta, true, err
	}
	if report.ResumeReady {
		return ModeResume, meta, true, nil
	}
	if report.RunnerExited != nil && !*report.RunnerExited {
		return ModeSend, meta, true, nil
	}
	return ModeRun, meta, true, nil
}

// TTYStartedInfo is passed to OnTTYStarted / OnTTYRestarted lifecycle hooks.
// Empty strings are OK when unknown; SessionID matches the run session when known.
// Reason is eventbus.ReasonTTYNew | ReasonTTYFollowup | ReasonTTYResume.
type TTYStartedInfo struct {
	SessionID string
	Runner    string
	Workspace string
	Reason    string
}

// Opts drives AutoSendOrResume. NewTerminal=false is the in-process P1 path.
// Dispatch hooks, when set, replace production send/run/resume for unit tests
// and prove no agent-run binary LookPath is required.
type Opts struct {
	SessionID         string
	Prompt            string
	WorkspaceDir      string
	AgentRunner       string
	AgentRunnerBinary string
	RunnerConfigHome  string
	Model             string
	// ModelReasoningEffort is optional Codex -c model_reasoning_effort=<level>.
	// Plumbed into agentui/agenttty RunOptions for codex-tty (ApplyCodexReasoningEffort).
	ModelReasoningEffort          string
	Open                          bool
	Detach                        bool
	NoSubmit                      bool
	KeepTTY                       bool
	JSON                          bool
	AllowRelocateResumeSessionDir bool
	// Driver is the host re-exec config for TTY __serve_* and should match
	// ForceNew FollowUp Driver (see pkgs/agentdriver). Zero → DefaultSelf in serve.
	Driver agentdriver.Driver
	// NewTerminal: P1 unit leaves keep false. When true, ForceNew may remain CLI-owned.
	NewTerminal  bool
	Env          []string
	PrependPaths []string
	// Color forces TTY child color env last (same policy as agent-run run --color).
	// Not persisted on meta; does not recolor agent-run own stdout/JSON.
	Color bool
	// ExitOnIdle / IdleTimeout are launch-time idle-exit flags (not persisted
	// on session meta). Same normalize/emit rules as FollowUpOpts.
	ExitOnIdle  bool
	IdleTimeout time.Duration
	Store       agentstorage.Store
	Stdout      io.Writer
	Stderr      io.Writer
	Probe       ProbeFunc

	// OnTTYStarted, when non-nil, is invoked once when AutoSendOrResume newly
	// establishes a live TTY for the session (successful ModeRun / first open).
	// Reason is "new". Nil = no-op. Not invoked on ModeSend or ModeResume.
	// Best-effort: panics in the hook are recovered and do not fail the run.
	OnTTYStarted func(info TTYStartedInfo)

	// OnTTYRestarted, when non-nil, is invoked after successful ModeSend
	// (follow-up into live TTY, reason "followup") or ModeResume (reason "resume").
	// Nil = no-op. Not invoked on ModeRun. Panics recovered like OnTTYStarted.
	OnTTYRestarted func(info TTYStartedInfo)

	// Optional dispatch overrides (nil = production path).
	RunSession    func(ctx context.Context, opts Opts, meta agentstorage.SessionMeta, found bool) error
	SendLive      func(ctx context.Context, opts Opts, meta agentstorage.SessionMeta) error
	ResumeSession func(ctx context.Context, opts Opts, meta agentstorage.SessionMeta) error
}

// AutoSendOrResume validates, classifies, and dispatches run|send|resume.
func AutoSendOrResume(ctx context.Context, opts Opts) error {
	if ctx == nil {
		ctx = context.Background()
	}
	sessionID := strings.TrimSpace(opts.SessionID)
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}
	opts.SessionID = sessionID

	if opts.Open && opts.Detach {
		return fmt.Errorf("--detach and --open are mutually exclusive; cannot use both")
	}

	if opts.Store == nil {
		return fmt.Errorf("store is required")
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	// Open-mode codex never binds runner_session_id before attach returns.
	// After /exit, bind from zombie scrollback footer and/or CODEX_HOME rollout
	// so Classify can ModeResume. EnsureCodexRunnerBound covers discovery +
	// live scrollback; zombie path remains gated (exit footer) for footer-only.
	if meta, found, rerr := resolveSession(opts.Store, sessionID); rerr == nil && found {
		meta = tryBindRunnerSessionFromZombie(opts.Store, meta)
		_, _ = EnsureCodexRunnerBound(opts.Store, meta, productionCodexBindOpts(opts.Store, meta))
	}

	// Prefer explicit opts.Probe; nil → LifecycleProbe inside Classify.
	mode, meta, found, err := Classify(opts.Store, sessionID, opts.Probe)
	if err != nil {
		return err
	}

	switch mode {
	case ModeSend:
		// Live follow-up into an existing TTY: never fire OnTTYStarted.
		var sendErr error
		if opts.SendLive != nil {
			sendErr = opts.SendLive(ctx, opts, meta)
		} else {
			sendErr = defaultSendLive(ctx, opts, meta)
		}
		if sendErr != nil {
			return sendErr
		}
		fireOnTTYRestarted(opts, meta, "followup")
		return nil
	case ModeResume:
		// Reclaim keep-alive zombie before ForceNew/child re-reserves the id.
		// (CLI resumeExistingSession also reclaims; this covers library + ModeRun
		// fall-throughs and parent openInNewTerminal → child paths.)
		ReclaimZombieTerminalIDs(opts.Store.Home(), effectiveRunner(opts, meta),
			meta.TerminalSessionID, meta.SessionID, opts.SessionID)
		var resumeErr error
		if opts.ResumeSession != nil {
			resumeErr = opts.ResumeSession(ctx, opts, meta)
		} else {
			resumeErr = defaultResumeSession(ctx, opts, meta)
		}
		if resumeErr != nil {
			return resumeErr
		}
		fireOnTTYRestarted(opts, meta, "resume")
		return nil
	default: // ModeRun
		// Found session re-open (unbound after exit, or first run): free zombie id.
		if found {
			ReclaimZombieTerminalIDs(opts.Store.Home(), effectiveRunner(opts, meta),
				meta.TerminalSessionID, meta.SessionID, opts.SessionID)
		}
		var runErr error
		if opts.RunSession != nil {
			runErr = opts.RunSession(ctx, opts, meta, found)
		} else {
			runErr = defaultRunSession(ctx, opts, meta, found)
		}
		if runErr != nil {
			return runErr
		}
		// First live TTY established for this session — fire once (best-effort).
		fireOnTTYStarted(opts, meta)
		return nil
	}
}

// fireOnTTYStarted invokes opts.OnTTYStarted once after a successful ModeRun
// dispatch. Nil hook is a no-op. Panics are recovered so bus side effects never
// fail the open path.
func fireOnTTYStarted(opts Opts, meta agentstorage.SessionMeta) {
	if opts.OnTTYStarted == nil {
		return
	}
	info := ttyLifecycleInfo(opts, meta, "new")
	defer recoverTTYHook(opts, "OnTTYStarted")
	opts.OnTTYStarted(info)
}

// fireOnTTYRestarted invokes opts.OnTTYRestarted after successful ModeSend/ModeResume.
func fireOnTTYRestarted(opts Opts, meta agentstorage.SessionMeta, reason string) {
	if opts.OnTTYRestarted == nil {
		return
	}
	info := ttyLifecycleInfo(opts, meta, reason)
	defer recoverTTYHook(opts, "OnTTYRestarted")
	opts.OnTTYRestarted(info)
}

func ttyLifecycleInfo(opts Opts, meta agentstorage.SessionMeta, reason string) TTYStartedInfo {
	info := TTYStartedInfo{
		SessionID: opts.SessionID,
		Runner:    effectiveRunner(opts, meta),
		Workspace: strings.TrimSpace(opts.WorkspaceDir),
		Reason:    reason,
	}
	if info.Workspace == "" {
		info.Workspace = strings.TrimSpace(meta.Workspace)
	}
	return info
}

func recoverTTYHook(opts Opts, name string) {
	if r := recover(); r != nil {
		w := opts.Stderr
		if w == nil {
			w = os.Stderr
		}
		fmt.Fprintf(w, "warning: %s panic: %v\n", name, r)
	}
}

func effectiveRunner(opts Opts, meta agentstorage.SessionMeta) string {
	if r := strings.TrimSpace(opts.AgentRunner); r != "" {
		return r
	}
	if r := strings.TrimSpace(meta.Runner); r != "" {
		return r
	}
	return "grok-tty"
}

// resolveSession looks up a bare session id. Missing → found=false (MODE=run).
// Ambiguous / other store errors are returned as err.
func resolveSession(store agentstorage.Store, sessionID string) (agentstorage.SessionMeta, bool, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return agentstorage.SessionMeta{}, false, nil
	}
	if strings.Contains(sessionID, "/") {
		return agentstorage.SessionMeta{}, false, fmt.Errorf("invalid session reference %q: expected bare session_id (not runner/session_id)", sessionID)
	}
	sess, err := store.GetSession(sessionID)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "session not found") {
			return agentstorage.SessionMeta{}, false, nil
		}
		return agentstorage.SessionMeta{}, false, err
	}
	return sess.Meta, true, nil
}
