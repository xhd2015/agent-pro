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

// Opts drives AutoSendOrResume. NewTerminal=false is the in-process P1 path.
// Dispatch hooks, when set, replace production send/run/resume for unit tests
// and prove no agent-run binary LookPath is required.
type Opts struct {
	SessionID                     string
	Prompt                        string
	WorkspaceDir                  string
	AgentRunner                   string
	AgentRunnerBinary             string
	RunnerConfigHome              string
	Model                         string
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
	NewTerminal bool
	Env         []string
	PrependPaths []string
	// Color forces TTY child color env last (same policy as agent-run run --color).
	// Not persisted on meta; does not recolor agent-run own stdout/JSON.
	Color  bool
	Store  agentstorage.Store
	Stdout       io.Writer
	Stderr       io.Writer
	Probe        ProbeFunc

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
	// After /exit, bind from zombie scrollback footer so Classify can ModeResume.
	if meta, found, rerr := resolveSession(opts.Store, sessionID); rerr == nil && found {
		_ = tryBindRunnerSessionFromZombie(opts.Store, meta)
	}

	// Prefer explicit opts.Probe; nil → LifecycleProbe inside Classify.
	mode, meta, found, err := Classify(opts.Store, sessionID, opts.Probe)
	if err != nil {
		return err
	}

	switch mode {
	case ModeSend:
		if opts.SendLive != nil {
			return opts.SendLive(ctx, opts, meta)
		}
		return defaultSendLive(ctx, opts, meta)
	case ModeResume:
		// Reclaim keep-alive zombie before ForceNew/child re-reserves the id.
		// (CLI resumeExistingSession also reclaims; this covers library + ModeRun
		// fall-throughs and parent openInNewTerminal → child paths.)
		ReclaimZombieTerminalIDs(opts.Store.Home(), effectiveRunner(opts, meta),
			meta.TerminalSessionID, meta.SessionID, opts.SessionID)
		if opts.ResumeSession != nil {
			return opts.ResumeSession(ctx, opts, meta)
		}
		return defaultResumeSession(ctx, opts, meta)
	default: // ModeRun
		// Found session re-open (unbound after exit, or first run): free zombie id.
		if found {
			ReclaimZombieTerminalIDs(opts.Store.Home(), effectiveRunner(opts, meta),
				meta.TerminalSessionID, meta.SessionID, opts.SessionID)
		}
		if opts.RunSession != nil {
			return opts.RunSession(ctx, opts, meta, found)
		}
		return defaultRunSession(ctx, opts, meta, found)
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
