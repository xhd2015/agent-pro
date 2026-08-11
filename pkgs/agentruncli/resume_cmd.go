package agentruncli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/agentui"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
	"github.com/xhd2015/less-gen/flags"
)

const resumeHelp = `
Usage: agent-run resume [OPTIONS] <session-id> ["followup…"]
       agent-run resume [OPTIONS] --grok-session-id ID ["followup…"]

Resume a finished agent-run session by re-invoking the runner with the
bound provider session id (grok --resume <runner_session_id>).

Resume is a shortcut of run: same flag family, but requires the session
to be bound and the runner to have exited. If the runner is still live,
use send instead.

Arguments:
  <session-id>   agent-run bare session id (mutually exclusive with --grok-session-id)
  followup       optional follow-up prompt (resume + send); omit to only resume

Options:
  --json              stream NDJSON AgentEvent lines to stdout
  --model MODEL       model name
  --keep-tty          keep TTY session alive after run completes
  --open              open keep-alive TTY and attach interactively
  --detach            start keep-alive TTY daemon and exit after registry (+ soft grok bind);
                      prints session-id and terminal-id on stdout; no attach / no stream;
                      exclusive with --open and --json; TTY only
  --no-submit         with --open: inject prompt without trailing Enter
  --dir DIR           workspace directory (default: session workspace or cwd)
  --allow-relocate-resume-session-dir
                      when --dir differs from grok session cwd, relocate the
                      grok session and continue (grok-tty only)
  --grok-session-id ID
                      resolve session by provider runner_session_id (meta.runner grok|grok-tty);
                      mutually exclusive with positional <session-id>
  --fork              branch via grok --fork-session into a NEW agent-run session
                      (parent Grok id from bound runner_session_id; skips live/exited gate)
  --agent-runner RUNNER   override runner (default: from session meta)
  --agent-runner-binary SPEC
                      agent executable: bare name/path or "binary flags..."
  --agent-runner-config-home PATH
                      agent data directory (grok: GROK_HOME, codex: CODEX_HOME);
                      replaces stored value when set
  --prepend-path DIR  append DIR to stored TTY child PATH prefixes (repeatable; TTY only)
  -e, --env KEY=VALUE append env var for the TTY agent runner child (repeatable; TTY only)
  -h, --help          show help
`

// resumeRunConfig holds CLI options for the resume path (subcommand or auto).
type resumeRunConfig struct {
	jsonFlag                      bool
	model                         string
	agentRunner                   string
	agentRunnerBinary             string
	agentRunnerConfigHome         string
	prependPaths                  []string // CLI appends only (absolute); merged with meta in resumeExistingSession
	envEntries                    []string // CLI appends only
	color                         bool     // not persisted; CLI-only for this resume invocation
	keepTTY                       bool
	openFlag                      bool
	detachFlag                    bool
	noSubmit                      bool
	dir                           string
	allowRelocateResumeSessionDir bool
	prompt                        string
	defaultRunner                 string
	fork                          bool
}

func runResume(args []string, defaultRunner string) error {
	var jsonFlag bool
	var model string
	var agentRunner string
	var agentRunnerBinary string
	var agentRunnerConfigHome string
	var prependPaths []string
	var envEntries []string
	var keepTTY bool
	var openFlag bool
	var detachFlag bool
	var noSubmit bool
	var dir string
	var allowRelocateResumeSessionDir bool
	var forkFlag bool
	var grokSessionID *string
	remaining, err := flags.Bool("--json", &jsonFlag).
		String("--model", &model).
		Bool("--keep-tty", &keepTTY).
		Bool("--open", &openFlag).
		Bool("--detach", &detachFlag).
		Bool("--no-submit", &noSubmit).
		String("--dir", &dir).
		Bool("--allow-relocate-resume-session-dir", &allowRelocateResumeSessionDir).
		Bool("--fork", &forkFlag).
		String("--grok-session-id", &grokSessionID).
		String("--agent-runner", &agentRunner).
		String("--agent-runner-binary", &agentRunnerBinary).
		String("--agent-runner-config-home", &agentRunnerConfigHome).
		StringSlice("--prepend-path", &prependPaths).
		StringSlice("-e,--env", &envEntries).
		Help("-h,--help", resumeHelp).
		Parse(args)
	if err != nil {
		return err
	}

	// With --grok-session-id, all remaining args are the optional followup prompt.
	// With positional <session-id>, remaining[0] is the id and the rest is followup.
	var sessionRef string
	var prompt string
	if grokSessionID != nil {
		prompt = strings.TrimSpace(strings.Join(remaining, " "))
	} else {
		if len(remaining) == 0 {
			return fmt.Errorf("resume requires <session-id> or --grok-session-id")
		}
		sessionRef = strings.TrimSpace(remaining[0])
		prompt = strings.TrimSpace(strings.Join(remaining[1:], " "))
	}

	if err := validateEnvFlags(envEntries); err != nil {
		return err
	}
	absPrepend, err := resolvePrependPaths(prependPaths)
	if err != nil {
		return err
	}
	absConfigHome, err := resolveAgentRunnerConfigHomeAbs(agentRunnerConfigHome)
	if err != nil {
		return err
	}
	envEntries = normalizeEnvEntries(envEntries)

	if detachFlag && openFlag {
		return fmt.Errorf("--detach and --open are mutually exclusive; cannot use both")
	}
	if detachFlag && jsonFlag {
		return fmt.Errorf("--detach and --json are mutually exclusive; cannot use both")
	}
	if openFlag && jsonFlag {
		return fmt.Errorf("--open and --json are mutually exclusive; cannot use both")
	}
	if noSubmit && !openFlag {
		return fmt.Errorf("--no-submit requires --open")
	}

	store, err := openStore()
	if err != nil {
		return err
	}
	meta, err := resolveSessionRef(store, sessionRef, grokSessionID)
	if err != nil {
		return err
	}

	return resumeExistingSession(store, meta, resumeRunConfig{
		jsonFlag:                      jsonFlag,
		model:                         model,
		agentRunner:                   agentRunner,
		agentRunnerBinary:             agentRunnerBinary,
		agentRunnerConfigHome:         absConfigHome,
		prependPaths:                  absPrepend,
		envEntries:                    envEntries,
		keepTTY:                       keepTTY,
		openFlag:                      openFlag,
		detachFlag:                    detachFlag,
		noSubmit:                      noSubmit,
		dir:                           dir,
		allowRelocateResumeSessionDir: allowRelocateResumeSessionDir,
		prompt:                        prompt,
		defaultRunner:                 defaultRunner,
		fork:                          forkFlag,
	})
}

// resumeExistingSession reclaims a zombie terminal if needed and re-invokes
// the provider with --resume <runner_session_id>. Workspace priority:
// --dir > meta.workspace > process cwd (+ stderr warning).
// With cfg.fork, creates a NEW agent-run session and launches grok --fork-session
// from the parent runner_session_id (does not require the parent runner to have exited).
func resumeExistingSession(store agentstorage.Store, meta agentstorage.SessionMeta, cfg resumeRunConfig) error {
	prompt := strings.TrimSpace(cfg.prompt)
	keepTTY := cfg.keepTTY
	openFlag := cfg.openFlag
	detachFlag := cfg.detachFlag
	// Empty followup is allowed: resume reopens the provider session
	// (grok --resume <id>) without sending a new turn. A followup is
	// resume + inject (like send after reopen). Without --open/--detach, keep the
	// TTY alive so the session can be attached/sent to after resume.
	if prompt == "" && !openFlag && !detachFlag {
		keepTTY = true
	}
	if detachFlag && openFlag {
		return fmt.Errorf("--detach and --open are mutually exclusive; cannot use both")
	}
	if detachFlag && cfg.jsonFlag {
		return fmt.Errorf("--detach and --json are mutually exclusive; cannot use both")
	}
	if openFlag && cfg.jsonFlag {
		return fmt.Errorf("--open and --json are mutually exclusive; cannot use both")
	}
	if cfg.noSubmit && !openFlag {
		return fmt.Errorf("--no-submit requires --open")
	}

	parentGrokID := strings.TrimSpace(meta.RunnerSessionID)
	if parentGrokID == "" {
		return fmt.Errorf("runner session not bound (missing runner_session_id); cannot resume")
	}

	// Gate: bound + exited (Resume.Ready). --fork branches from on-disk Grok state
	// and does not require the parent agent-run TTY to have exited.
	if !cfg.fork {
		report := probeSessionStatus(store, meta)
		if report.Runner.Status != "bound" {
			return fmt.Errorf("runner session not bound (missing runner_session_id); cannot resume")
		}
		if report.Runner.Exited == nil || !*report.Runner.Exited {
			return fmt.Errorf("cannot resume: runner not exited (still active/live); use send instead of resume")
		}
	}

	runner := strings.TrimSpace(cfg.agentRunner)
	if runner == "" {
		runner = strings.TrimSpace(meta.Runner)
	}
	if runner == "" {
		runner = cfg.defaultRunner
	}
	if err := validateRunner(runner); err != nil {
		return err
	}
	if cfg.fork {
		r := strings.TrimSpace(runner)
		if r != "" && r != "grok-tty" && r != "grok" {
			return fmt.Errorf("--fork requires grok-tty (got %s)", r)
		}
		runner = "grok-tty"
	}
	if openFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--open requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	if detachFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--detach requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	// CLI session-env flags only: stored values reapply without re-checking TTY.
	if err := requireTTYForSessionEnv(runner, cfg.prependPaths, cfg.envEntries); err != nil {
		return err
	}
	if err := requireTTYForColor(runner, cfg.color); err != nil {
		return err
	}

	// Resume merge: stored paths/env always applied; CLI flags append only.
	// Config home: CLI replaces scalar when set, else use stored.
	effectivePrepend := appendStringLists(meta.PrependPaths, cfg.prependPaths)
	effectiveEnv := appendStringLists(meta.Env, cfg.envEntries)
	configHome := strings.TrimSpace(cfg.agentRunnerConfigHome)
	if configHome == "" {
		configHome = strings.TrimSpace(meta.AgentRunnerConfigHome)
	}

	workspace, err := resolveResumeWorkspace(cfg.dir, meta)
	if err != nil {
		return err
	}
	// When --dir is set and runner is grok-tty, compare to Grok session info.cwd.
	// Mismatch without --allow-relocate-resume-session-dir is an error; with the
	// flag, relocate the Grok session and update meta.workspace.
	if strings.TrimSpace(cfg.dir) != "" {
		if err := ensureGrokResumeDirMatchesCWD(store, meta, runner, cfg.dir, workspace, configHome, cfg.allowRelocateResumeSessionDir); err != nil {
			return err
		}
	}
	model := strings.TrimSpace(cfg.model)
	if model == "" {
		model = strings.TrimSpace(meta.Model)
	}

	// --fork: new agent-run session pre-bound to parent Grok id; launch with --fork-session.
	if cfg.fork {
		newID, genErr := generateAutoSessionID(prompt, runner, store.Home())
		if genErr != nil {
			return genErr
		}
		if _, err := store.GetSession(newID); err == nil {
			return fmt.Errorf("session already exists: %s", newID)
		}
		createMeta := agentstorage.SessionMeta{
			Runner:                runner,
			SessionID:             newID,
			Status:                "running",
			Model:                 model,
			InitialPrompt:         prompt,
			RunnerSessionID:       parentGrokID,
			Workspace:             workspace,
			PrependPaths:          append([]string(nil), effectivePrepend...),
			Env:                   append([]string(nil), effectiveEnv...),
			AgentRunnerConfigHome: configHome,
		}
		if err := store.CreateSession(newID, createMeta); err != nil {
			return err
		}
		return agentui.Run(context.Background(), agentui.RunOptions{
			Prompt:                prompt,
			Runner:                runner,
			Model:                 model,
			SessionID:             newID,
			AgentRunnerBinary:     cfg.agentRunnerBinary,
			AgentRunnerConfigHome: configHome,
			PrependPaths:          effectivePrepend,
			Env:                   effectiveEnv,
			Color:                 cfg.color,
			JSON:                  cfg.jsonFlag,
			Workspace:             workspace,
			KeepTerminalAlive:     keepAliveOpenDetach(keepTTY, openFlag, detachFlag),
			Open:                  openFlag,
			Detach:                detachFlag,
			NoSubmit:              cfg.noSubmit,
			Fork:                  true,
			Driver:                mergeHostDriver(agentdriver.Driver{}),
			Store:                 store,
			Stdout:                os.Stdout,
			Stderr:                os.Stderr,
		})
	}

	// Zombie keep-alive often still holds the TTY registry id (same as agent
	// session id when --session-id was used). Reclaim before agentui.Run so
	// ReserveCustomSessionID can reuse it; if reclaim cannot free the id,
	// fall back to auto session-N while keeping agent storage SessionID.
	ttySessionID, preferAuto, err := prepareResumeTerminalID(store, runner, meta)
	if err != nil {
		return err
	}

	return agentui.Run(context.Background(), agentui.RunOptions{
		Prompt:                prompt,
		Runner:                runner,
		Model:                 model,
		SessionID:             meta.SessionID,
		TerminalSessionID:     ttySessionID,
		PreferAutoTerminal:    preferAuto,
		AgentRunnerBinary:     cfg.agentRunnerBinary,
		AgentRunnerConfigHome: configHome,
		PrependPaths:          effectivePrepend,
		Env:                   effectiveEnv,
		Color:                 cfg.color,
		JSON:                  cfg.jsonFlag,
		Workspace:             workspace,
		KeepTerminalAlive:     keepAliveOpenDetach(keepTTY, openFlag, detachFlag),
		Open:                  openFlag,
		Detach:                detachFlag,
		NoSubmit:              cfg.noSubmit,
		Driver:                mergeHostDriver(agentdriver.Driver{}),
		Store:                 store,
		Stdout:                os.Stdout,
		Stderr:                os.Stderr,
	})
}

// ensureGrokResumeDirMatchesCWD compares resolved --dir to the Grok session's
// info.cwd for meta.runner_session_id. Non-grok-tty runners skip the check.
// On mismatch: error unless allowRelocate, then RelocateCWD + update meta.workspace.
//
// dirFlag is the raw --dir value; resolvedDir is Abs+EvalSymlinks from resolveRunDir.
// RelocateCWD receives Abs(dirFlag) without EvalSymlinks so Grok session keys
// match url.PathEscape(filepath.Abs(cwd)) layout used by Grok/fixtures.
func ensureGrokResumeDirMatchesCWD(store agentstorage.Store, meta agentstorage.SessionMeta, runner, dirFlag, resolvedDir, configHome string, allowRelocate bool) error {
	runner = strings.TrimSpace(runner)
	if runner != "grok-tty" {
		return nil
	}
	compareDir := strings.TrimSpace(resolvedDir)
	if compareDir == "" {
		compareDir = strings.TrimSpace(dirFlag)
	}
	if compareDir == "" {
		return nil
	}
	runnerSessionID := strings.TrimSpace(meta.RunnerSessionID)
	if runnerSessionID == "" {
		return fmt.Errorf("runner session not bound (missing runner_session_id); cannot compare grok session cwd")
	}
	grokHome := agenttty.GrokHomeForRunner(configHome)
	sess, err := sessions.Find(grokHome, runnerSessionID)
	if err != nil {
		// Session not present under Grok home (common when fixtures omit Grok
		// layout and only exercise agent-run meta.workspace / --dir). Skip the
		// cwd gate so resume continues as before.
		return nil
	}
	grokCWD := strings.TrimSpace(sess.CWD)
	if grokCWD == "" {
		// Found session but no cwd to compare — cannot gate safely.
		return nil
	}
	if canonicalPathsEqual(compareDir, grokCWD) {
		return nil
	}
	// Abs without EvalSymlinks — matches sessions.RelocateCWD / Grok layout encoding.
	relocateTarget := compareDir
	if abs, absErr := filepath.Abs(strings.TrimSpace(dirFlag)); absErr == nil && abs != "" {
		relocateTarget = abs
	}
	if !allowRelocate {
		return fmt.Errorf("--dir %s differs from grok session cwd %s (session %s); refusing to resume with a different workspace; pass --allow-relocate-resume-session-dir to relocate the grok session and continue",
			relocateTarget, grokCWD, runnerSessionID)
	}
	fmt.Fprintf(os.Stderr, "warning: relocating grok session %s cwd from %s to %s\n", runnerSessionID, grokCWD, relocateTarget)
	if _, err := sessions.RelocateCWD(runnerSessionID, relocateTarget, &sessions.RelocateCWDOptions{GrokHome: grokHome}); err != nil {
		return fmt.Errorf("relocate grok session cwd: %w", err)
	}
	if err := store.UpdateSessionWorkspace(meta.SessionID, relocateTarget); err != nil {
		return fmt.Errorf("update session workspace after relocate: %w", err)
	}
	return nil
}

// canonicalPathsEqual reports whether a and b refer to the same directory path
// after Abs, EvalSymlinks, Clean, and /private prefix normalization.
func canonicalPathsEqual(a, b string) bool {
	na := normalizePathForCompare(a)
	nb := normalizePathForCompare(b)
	if na == "" || nb == "" {
		return na == nb
	}
	if na == nb {
		return true
	}
	// os.SameFile when both exist (covers remaining symlink /private cases).
	infoA, errA := os.Stat(na)
	infoB, errB := os.Stat(nb)
	if errA == nil && errB == nil {
		return os.SameFile(infoA, infoB)
	}
	return false
}

func normalizePathForCompare(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return filepath.Clean(p)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)
	// macOS: /var → /private/var after EvalSymlinks; also accept the reverse.
	if strings.HasPrefix(abs, "/private/") {
		alt := strings.TrimPrefix(abs, "/private")
		if alt != "" && alt[0] == '/' {
			if a, errA := os.Lstat(abs); errA == nil {
				if b, errB := os.Lstat(alt); errB == nil && os.SameFile(a, b) {
					return filepath.Clean(alt)
				}
			}
			// Prefer stripped form for string equality when both styles encode the same path.
			return filepath.Clean(alt)
		}
	}
	return abs
}

// resolveResumeWorkspace picks the provider cwd for resume:
// 1) --dir if set (validated), 2) meta.workspace (must exist as a directory),
// 3) empty → process cwd (+ warn). Does not fall back to process cwd when
// meta.workspace is set but missing or not a directory.
func resolveResumeWorkspace(dir string, meta agentstorage.SessionMeta) (string, error) {
	workspace, err := resolveRunDir(dir)
	if err != nil {
		return "", err
	}
	if workspace != "" {
		return workspace, nil
	}
	workspace = strings.TrimSpace(meta.Workspace)
	if workspace != "" {
		abs, err := filepath.Abs(workspace)
		if err != nil {
			return "", fmt.Errorf("session workspace: %w", err)
		}
		if resolved, err := filepath.EvalSymlinks(abs); err == nil {
			abs = resolved
		}
		info, err := os.Stat(abs)
		if err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("session workspace does not exist: %s; pass --dir <existing-directory> to override", abs)
			}
			return "", fmt.Errorf("session workspace: %w", err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("session workspace is not a directory: %s; pass --dir <existing-directory> to override", abs)
		}
		return abs, nil
	}
	fmt.Fprintln(os.Stderr, "warning: session has no workspace recorded; using process cwd")
	return "", nil
}

// prepareResumeTerminalID reclaims a zombie-held terminal registry id so resume
// can re-reserve it. Primary: free and reuse the same id (meta.terminal_session_id
// or SessionID). Fallback: PreferAutoTerminal when the id remains held.
//
// Live agents are not reclaimed here — resume already gates on exited=true.
func prepareResumeTerminalID(store agentstorage.Store, runner string, meta agentstorage.SessionMeta) (ttySessionID string, preferAuto bool, err error) {
	// Prefer explicit terminal id; fall back to agent session id (production
	// --session-id made them equal and is what ReserveCustomSessionID reuses).
	termID := strings.TrimSpace(meta.TerminalSessionID)
	if termID == "" {
		termID = strings.TrimSpace(meta.SessionID)
	}
	sessionID := strings.TrimSpace(meta.SessionID)

	// Ids that may hold the registry entry and/or will be reserved as TTY id.
	// Always attempt reclaim for both when distinct.
	ids := []string{}
	if termID != "" {
		ids = append(ids, termID)
	}
	if sessionID != "" && sessionID != termID {
		ids = append(ids, sessionID)
	}

	cfg := registryConfigForRunner(store.Home(), runner)
	for _, id := range ids {
		if !ttywatch.SessionIDInUse(cfg, id) {
			// Also try other provider registry subdirs when primary is free.
			_ = reclaimAcrossProviders(store.Home(), id)
			continue
		}
		if reclaimErr := ttywatch.ReclaimSessionID(cfg, id); reclaimErr != nil {
			// Best-effort; still try cross-provider + fallback below.
			_ = reclaimErr
		}
		// Always also reclaim under any provider subdir that holds the id.
		_ = reclaimAcrossProviders(store.Home(), id)
	}

	// Desired TTY id after reclaim: prefer termID (same as agent when matched).
	desired := termID
	if desired == "" {
		desired = sessionID
	}
	if desired == "" {
		return "", false, nil
	}

	// Re-check the registry cfg that agentui/agenttty will use for this runner.
	if !ttywatch.SessionIDInUse(cfg, desired) && !sessionIDHeldAnywhere(store.Home(), desired) {
		// Persist desired terminal id when meta lacked it but we will reuse SessionID.
		if strings.TrimSpace(meta.TerminalSessionID) == "" && desired == sessionID {
			_ = store.UpdateSessionTerminalSessionID(sessionID, desired)
		}
		return desired, false, nil
	}

	// Fallback B: allocate new terminal id (auto session-N via PreferAutoTerminal).
	// Clear the stale terminal mapping; agentui will persist the new id on start.
	if sessionID != "" {
		_ = store.UpdateSessionTerminalSessionID(sessionID, "")
	}
	return "", true, nil
}

func registryConfigForRunner(home, runner string) ttywatch.RegistryConfig {
	subdir := ""
	if p, ok := agenttty.Get(runner); ok {
		subdir = p.RegistryDir
	}
	if subdir == "" {
		if runner == "" {
			runner = "grok-tty"
		}
		subdir = runner + "-registry"
	}
	return ttywatch.RegistryConfig{Home: home, Subdir: subdir}
}

// reclaimAcrossProviders force-reclaims id under every registered TTY provider
// registry subdir (and common agent-run layouts).
func reclaimAcrossProviders(home, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}
	seen := map[string]bool{}
	for _, p := range agenttty.ProviderListSorted() {
		subdir := p.RegistryDir
		if subdir == "" || seen[subdir] {
			continue
		}
		seen[subdir] = true
		cfg := ttywatch.RegistryConfig{Home: home, Subdir: subdir}
		if ttywatch.SessionIDInUse(cfg, sessionID) {
			_ = ttywatch.ReclaimSessionID(cfg, sessionID)
		}
	}
	return nil
}

func sessionIDHeldAnywhere(home, sessionID string) bool {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return false
	}
	for _, p := range agenttty.ProviderListSorted() {
		subdir := p.RegistryDir
		if subdir == "" {
			continue
		}
		cfg := ttywatch.RegistryConfig{Home: home, Subdir: subdir}
		if ttywatch.SessionIDInUse(cfg, sessionID) {
			return true
		}
	}
	return false
}
