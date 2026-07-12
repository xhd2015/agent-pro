package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/agentui"
	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
	"github.com/xhd2015/less-gen/flags"
)

const resumeHelp = `
Usage: agent-run resume [OPTIONS] <session-id> ["followup…"]

Resume a finished agent-run session by re-invoking the runner with the
bound provider session id (grok --resume <runner_session_id>).

Resume is a shortcut of run: same flag family, but requires the session
to be bound and the runner to have exited. If the runner is still live,
use send instead.

Arguments:
  <session-id>   agent-run session id (or runner/session_id)
  followup       optional follow-up prompt (resume + send); omit to only resume

Options:
  --json              stream NDJSON AgentEvent lines to stdout
  --model MODEL       model name
  --keep-tty          keep TTY session alive after run completes
  --open              open keep-alive TTY and attach interactively
  --no-submit         with --open: inject prompt without trailing Enter
  --dir DIR           workspace directory (default: session workspace or cwd)
  --agent-runner RUNNER   override runner (default: from session meta)
  --agent-runner-binary SPEC
                      agent executable: bare name/path or "binary flags..."
  --agent-runner-config-home PATH
                      agent data directory (grok: GROK_HOME, codex: CODEX_HOME)
  -h, --help          show help
`

// resumeRunConfig holds CLI options for the resume path (subcommand or auto).
type resumeRunConfig struct {
	jsonFlag              bool
	model                 string
	agentRunner           string
	agentRunnerBinary     string
	agentRunnerConfigHome string
	keepTTY               bool
	openFlag              bool
	noSubmit              bool
	dir                   string
	prompt                string
	defaultRunner         string
}

func runResume(args []string, defaultRunner string) error {
	var jsonFlag bool
	var model string
	var agentRunner string
	var agentRunnerBinary string
	var agentRunnerConfigHome string
	var keepTTY bool
	var openFlag bool
	var noSubmit bool
	var dir string
	remaining, err := flags.Bool("--json", &jsonFlag).
		String("--model", &model).
		Bool("--keep-tty", &keepTTY).
		Bool("--open", &openFlag).
		Bool("--no-submit", &noSubmit).
		String("--dir", &dir).
		String("--agent-runner", &agentRunner).
		String("--agent-runner-binary", &agentRunnerBinary).
		String("--agent-runner-config-home", &agentRunnerConfigHome).
		Help("-h,--help", resumeHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) == 0 {
		return fmt.Errorf("resume requires <session-id>")
	}
	sessionRef := strings.TrimSpace(remaining[0])
	prompt := strings.TrimSpace(strings.Join(remaining[1:], " "))

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
	meta, err := resolveSessionMeta(store, sessionRef)
	if err != nil {
		return err
	}

	return resumeExistingSession(store, meta, resumeRunConfig{
		jsonFlag:              jsonFlag,
		model:                 model,
		agentRunner:           agentRunner,
		agentRunnerBinary:     agentRunnerBinary,
		agentRunnerConfigHome: agentRunnerConfigHome,
		keepTTY:               keepTTY,
		openFlag:              openFlag,
		noSubmit:              noSubmit,
		dir:                   dir,
		prompt:                prompt,
		defaultRunner:         defaultRunner,
	})
}

// resumeExistingSession reclaims a zombie terminal if needed and re-invokes
// the provider with --resume <runner_session_id>. Workspace priority:
// --dir > meta.workspace > process cwd (+ stderr warning).
func resumeExistingSession(store agentstorage.Store, meta agentstorage.SessionMeta, cfg resumeRunConfig) error {
	prompt := strings.TrimSpace(cfg.prompt)
	keepTTY := cfg.keepTTY
	openFlag := cfg.openFlag
	// Empty followup is allowed: resume reopens the provider session
	// (grok --resume <id>) without sending a new turn. A followup is
	// resume + inject (like send after reopen). Without --open, keep the
	// TTY alive so the session can be attached/sent to after resume.
	if prompt == "" && !openFlag {
		keepTTY = true
	}
	if openFlag && cfg.jsonFlag {
		return fmt.Errorf("--open and --json are mutually exclusive; cannot use both")
	}
	if cfg.noSubmit && !openFlag {
		return fmt.Errorf("--no-submit requires --open")
	}

	// Gate: bound + exited (Resume.Ready).
	report := probeSessionStatus(store, meta)
	if report.Runner.Status != "bound" || strings.TrimSpace(meta.RunnerSessionID) == "" {
		return fmt.Errorf("runner session not bound (missing runner_session_id); cannot resume")
	}
	if report.Runner.Exited == nil || !*report.Runner.Exited {
		return fmt.Errorf("cannot resume: runner not exited (still active/live); use send instead of resume")
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
	if openFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--open requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}

	workspace, err := resolveResumeWorkspace(cfg.dir, meta)
	if err != nil {
		return err
	}
	model := strings.TrimSpace(cfg.model)
	if model == "" {
		model = strings.TrimSpace(meta.Model)
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
		AgentRunnerConfigHome: cfg.agentRunnerConfigHome,
		JSON:                  cfg.jsonFlag,
		Workspace:             workspace,
		KeepTerminalAlive:     keepTTY || openFlag,
		Open:                  openFlag,
		NoSubmit:              cfg.noSubmit,
		Store:                 store,
		Stdout:                os.Stdout,
		Stderr:                os.Stderr,
	})
}

// resolveResumeWorkspace picks the provider cwd for resume:
// 1) --dir if set (validated), 2) meta.workspace, 3) empty → process cwd (+ warn).
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
		return workspace, nil
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
			_ = store.UpdateSessionTerminalSessionID(runner, sessionID, desired)
		}
		return desired, false, nil
	}

	// Fallback B: allocate new terminal id (auto session-N via PreferAutoTerminal).
	// Clear the stale terminal mapping; agentui will persist the new id on start.
	if sessionID != "" {
		_ = store.UpdateSessionTerminalSessionID(runner, sessionID, "")
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
