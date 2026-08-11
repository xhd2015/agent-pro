package agentruncli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentsend"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/agentui"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	flags "github.com/xhd2015/less-flags"
)

const runHelp = `
Usage: agent-run run [OPTIONS] ["prompt"]

Options:
  --json              stream NDJSON AgentEvent lines to stdout
  --model MODEL       model name
  --model-reasoning-effort LEVEL  optional reasoning effort (pass-through; empty/omitted = no default)
  --session ID        session id
  --session-id ID     alias for --session
  --session-id-from-prompt   generate session id from prompt slug (storage + TTY registry)
  --auto-send-or-resume      with --session-id: live→send, exited+bound→resume, else→run
  --new-terminal      with --auto-send-or-resume: open a new iTerm2 window for run/resume
                      (ignored when MODE=send / live)
  --keep-tty          keep TTY session alive after run completes
  --open              open keep-alive TTY and attach interactively (silent until detach; prints session id after);
                      with --auto-send-or-resume: required shape for create/resume; ignored when live (send only)
  --detach            start keep-alive TTY daemon and exit after registry (+ soft grok bind);
                      prints session-id and terminal-id on stdout; no attach / no stream;
                      exclusive with --open and --json; TTY only; empty prompt OK
  --no-submit         with --open: inject prompt into TTY without trailing Enter (no auto-submit);
                      with --auto-send-or-resume send path: inject without Enter
  --dir DIR           workspace directory (default: process cwd; resume uses session workspace)
  --allow-relocate-resume-session-dir
                      when resume --dir differs from grok session cwd, relocate the
                      grok session and continue (grok-tty only; with --auto-send-or-resume)
  --resume-from-grok-session ID
                      import an external Grok CLI session by provider id (grok-tty only);
                      validates id / runner / GROK_HOME presence / not already mapped / --dir vs cwd
  --fork              with --resume-from-grok-session: branch via grok --fork-session into a
                      NEW agent-run session (skips already-mapped for the parent Grok id)
  --agent-runner RUNNER   codex, codex-tty, grok-tty, opencode, fake-codex, ...
  --agent-runner-binary SPEC
                      agent executable: bare name/path or shell-style "binary flags..."
  --agent-runner-config-home PATH
                      agent data directory (grok: GROK_HOME, codex: CODEX_HOME);
                      default: AGENT_RUNNER_CONFIG_HOME env
  --prepend-path DIR  prepend DIR to the TTY agent runner child PATH (repeatable; TTY only)
  -e, --env KEY=VALUE set env var on the TTY agent runner child process (repeatable; TTY only)
  --color             force color on the TTY agent runner child (unset NO_COLOR; FORCE_COLOR/CLICOLOR; TTY only)
  --event-bus-url URL publish agent.tty.started after a successful new-terminal open (best-effort)
  --event-bus-token TOKEN  optional Bearer token for event-bus publish
  -h, --help          show help
`

func runHeadless(args []string, defaultRunner string) error {
	var jsonFlag bool
	var model string
	var modelReasoningEffort string
	var sessionID string
	var sessionIDFromPrompt bool
	var autoSendOrResume bool
	var newTerminal bool
	var agentRunner string
	var agentRunnerBinary string
	var agentRunnerConfigHome string
	var prependPaths []string
	var envEntries []string
	var colorFlag bool
	var keepTTY bool
	var openFlag bool
	var detachFlag bool
	var noSubmit bool
	var dir string
	var allowRelocateResumeSessionDir bool
	var resumeFromGrokSession *string
	var forkFlag bool
	var eventBusURL string
	var eventBusToken string
	var recorded flags.Flags
	remaining, err := flags.Bool("--json", &jsonFlag).
		String("--model", &model).
		String("--model-reasoning-effort", &modelReasoningEffort).
		String("--session,--session-id", &sessionID).
		Bool("--session-id-from-prompt", &sessionIDFromPrompt).
		Bool("--auto-send-or-resume", &autoSendOrResume).
		Bool("--new-terminal", &newTerminal).
		Bool("--keep-tty", &keepTTY).
		Bool("--open", &openFlag).
		Bool("--detach", &detachFlag).
		Bool("--no-submit", &noSubmit).
		String("--dir", &dir).
		Bool("--allow-relocate-resume-session-dir", &allowRelocateResumeSessionDir).
		String("--resume-from-grok-session", &resumeFromGrokSession).
		Bool("--fork", &forkFlag).
		String("--agent-runner", &agentRunner).
		String("--agent-runner-binary", &agentRunnerBinary).
		String("--agent-runner-config-home", &agentRunnerConfigHome).
		StringSlice("--prepend-path", &prependPaths).
		StringSlice("-e,--env", &envEntries).
		Bool("--color", &colorFlag).
		String("--event-bus-url", &eventBusURL).
		String("--event-bus-token", &eventBusToken).
		Help("-h,--help", runHelp).
		CollectParsedFlags(&recorded).
		Parse(args)
	if err != nil {
		return err
	}
	prompt := strings.TrimSpace(strings.Join(remaining, " "))

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

	// Import external Grok session: P1 validation + P2 create/pre-bind + P3 open/detach.
	// Note: global `agent-run --agent-runner X run ...` strips the flag in main
	// into defaultRunner; subcommand `run --agent-runner X` is also stripped the
	// same way when the flag appears after `run` (main consumes it). Use both.
	if resumeFromGrokSession != nil {
		if autoSendOrResume {
			return fmt.Errorf("--resume-from-grok-session and --auto-send-or-resume are mutually exclusive; cannot use both")
		}
		return runResumeFromGrokSession(resumeFromGrokOpts{
			grokSessionID:     *resumeFromGrokSession,
			agentRunner:       agentRunner,
			defaultRunner:     defaultRunner,
			configHome:        absConfigHome,
			dir:               dir,
			sessionID:         sessionID,
			prompt:            prompt,
			model:             model,
			agentRunnerBinary: agentRunnerBinary,
			prependPaths:      absPrepend,
			envEntries:        envEntries,
			color:             colorFlag,
			jsonFlag:          jsonFlag,
			keepTTY:           keepTTY,
			openFlag:          openFlag,
			detachFlag:        detachFlag,
			noSubmit:          noSubmit,
			fork:              forkFlag,
		})
	}
	if forkFlag {
		return fmt.Errorf("--fork requires --resume-from-grok-session (or use: agent-run resume --fork …)")
	}

	if newTerminal && !autoSendOrResume {
		return fmt.Errorf("--new-terminal requires --auto-send-or-resume")
	}

	if autoSendOrResume {
		return runAutoSendOrResume(autoSendOrResumeOpts{
			jsonFlag:                      jsonFlag,
			model:                         model,
			modelReasoningEffort:          modelReasoningEffort,
			sessionID:                     sessionID,
			sessionIDFromPrompt:           sessionIDFromPrompt,
			agentRunner:                   agentRunner,
			agentRunnerBinary:             agentRunnerBinary,
			agentRunnerConfigHome:         absConfigHome,
			prependPaths:                  absPrepend,
			envEntries:                    envEntries,
			color:                         colorFlag,
			keepTTY:                       keepTTY,
			openFlag:                      openFlag,
			detachFlag:                    detachFlag,
			noSubmit:                      noSubmit,
			dir:                           dir,
			allowRelocateResumeSessionDir: allowRelocateResumeSessionDir,
			prompt:                        prompt,
			defaultRunner:                 defaultRunner,
			newTerminal:                   newTerminal,
			eventBusURL:                   eventBusURL,
			eventBusToken:                 eventBusToken,
			recorded:                      recorded,
		})
	}

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
	if prompt == "" && !openFlag && !detachFlag {
		return fmt.Errorf("prompt is required")
	}
	if sessionIDFromPrompt && strings.TrimSpace(sessionID) != "" {
		return fmt.Errorf("--session/--session-id and --session-id-from-prompt are mutually exclusive; cannot use both")
	}
	workspace, err := resolveRunDir(dir)
	if err != nil {
		return err
	}
	runner := resolveCLIRunner(agentRunner, defaultRunner)
	if err := validateRunner(runner); err != nil {
		return err
	}
	if openFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--open requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	if detachFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--detach requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	if err := requireTTYForSessionEnv(runner, absPrepend, envEntries); err != nil {
		return err
	}
	if err := requireTTYForColor(runner, colorFlag); err != nil {
		return err
	}
	store, err := openStore()
	if err != nil {
		return err
	}
	if sessionIDFromPrompt {
		id, genErr := generateAutoSessionID(prompt, runner, store.Home())
		if genErr != nil {
			return genErr
		}
		sessionID = id
	}
	return agentui.Run(context.Background(), agentui.RunOptions{
		Prompt:                prompt,
		Runner:                runner,
		Model:                 model,
		ModelReasoningEffort:  modelReasoningEffort,
		SessionID:             sessionID,
		AgentRunnerBinary:     agentRunnerBinary,
		AgentRunnerConfigHome: absConfigHome,
		PrependPaths:          absPrepend,
		Env:                   envEntries,
		Color:                 colorFlag,
		JSON:                  jsonFlag,
		Workspace:             workspace,
		KeepTerminalAlive:     keepAliveOpenDetach(keepTTY, openFlag, detachFlag),
		Open:                  openFlag,
		Detach:                detachFlag,
		NoSubmit:              noSubmit,
		Driver:                mergeHostDriver(agentdriver.Driver{}),
		Store:                 store,
		Stdout:                os.Stdout,
		Stderr:                os.Stderr,
	})
}

// keepAliveOpenDetach: detach always keeps; open keeps only when
// OpenCloseExits is off (AGENT_RUN_OPEN_CLOSE_EXITS=0) or --keep-tty.
func keepAliveOpenDetach(keepTTY, open, detach bool) bool {
	if keepTTY || detach {
		return true
	}
	if open && !agenttty.OpenCloseExits() {
		return true
	}
	return false
}

// resolveCLIRunner picks the agent runner for run/auto paths.
// Explicit --agent-runner wins, then global default, then grok-tty (agent-run's
// primary TTY runner). agentui still falls back to store config / opencode when
// Runner is left empty by other call sites.
func resolveCLIRunner(flagRunner, defaultRunner string) string {
	if r := strings.TrimSpace(flagRunner); r != "" {
		return r
	}
	if r := strings.TrimSpace(defaultRunner); r != "" {
		return r
	}
	return "grok-tty"
}

// resumeFromGrokOpts carries CLI inputs for run --resume-from-grok-session.
type resumeFromGrokOpts struct {
	grokSessionID     string
	agentRunner       string
	defaultRunner     string
	configHome        string
	dir               string
	sessionID         string
	prompt            string
	model             string
	agentRunnerBinary string
	prependPaths      []string
	envEntries        []string
	color             bool
	jsonFlag          bool
	keepTTY           bool
	openFlag          bool
	detachFlag        bool
	noSubmit          bool
	fork              bool
}

// runResumeFromGrokSession validates (P1), CreateSession-pre-binds the Grok UUID
// on a new agent-run session (P2), then launches via agentui.Run (headless / open /
// detach) so meta.RunnerSessionID drives ResumeSessionID → provider argv
// `--resume <uuid>`.
func runResumeFromGrokSession(opts resumeFromGrokOpts) error {
	id := strings.TrimSpace(opts.grokSessionID)
	if id == "" {
		return fmt.Errorf("--resume-from-grok-session requires a non-empty value")
	}
	// Same mutual-exclusion rules as normal run for open/detach/json/no-submit.
	if opts.detachFlag && opts.openFlag {
		return fmt.Errorf("--detach and --open are mutually exclusive; cannot use both")
	}
	if opts.detachFlag && opts.jsonFlag {
		return fmt.Errorf("--detach and --json are mutually exclusive; cannot use both")
	}
	if opts.openFlag && opts.jsonFlag {
		return fmt.Errorf("--open and --json are mutually exclusive; cannot use both")
	}
	if opts.noSubmit && !opts.openFlag {
		return fmt.Errorf("--no-submit requires --open")
	}

	// Prefer explicit run-flag runner, then global/default (main often strips
	// --agent-runner before the subcommand sees it). Omitted → grok-tty.
	runner := strings.TrimSpace(opts.agentRunner)
	if runner == "" {
		runner = strings.TrimSpace(opts.defaultRunner)
	}
	if runner == "" {
		runner = "grok-tty"
	}
	if runner != "grok-tty" {
		return fmt.Errorf("--resume-from-grok-session requires grok-tty (got %s); only grok-tty is allowed when --agent-runner is set", runner)
	}
	if opts.openFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--open requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	if opts.detachFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--detach requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	if err := requireTTYForSessionEnv(runner, opts.prependPaths, opts.envEntries); err != nil {
		return err
	}
	if err := requireTTYForColor(runner, opts.color); err != nil {
		return err
	}

	grokHome := agenttty.GrokHomeForRunner(opts.configHome)
	grokSess, err := sessions.Find(grokHome, id)
	if err != nil {
		return err
	}

	store, err := openStore()
	if err != nil {
		return err
	}
	list, err := store.ListSessions()
	if err != nil {
		return err
	}
	// Without --fork: 1:1 import identity — reject re-import of an already-bound Grok id.
	// With --fork: branch into a NEW agent-run session; parent may already be mapped.
	if !opts.fork {
		for _, m := range list {
			if !isGrokRunner(m.Runner) {
				continue
			}
			if strings.TrimSpace(m.RunnerSessionID) == id {
				return fmt.Errorf("grok session %s is already mapped to agent-run session %s; use resume --grok-session-id %s (or re-run with --fork to branch)",
					id, m.SessionID, id)
			}
		}
	}

	// Workspace: default Grok info.cwd; --dir allowed only when it matches (no relocate).
	workspace := strings.TrimSpace(grokSess.CWD)
	if workspace != "" {
		if abs, absErr := filepath.Abs(workspace); absErr == nil {
			workspace = abs
		}
	}
	if strings.TrimSpace(opts.dir) != "" {
		dirWS, err := resolveRunDir(opts.dir)
		if err != nil {
			return err
		}
		if workspace != "" && !canonicalPathsEqual(dirWS, workspace) {
			absDir := dirWS
			if a, absErr := filepath.Abs(strings.TrimSpace(opts.dir)); absErr == nil && a != "" {
				absDir = a
			}
			return fmt.Errorf("--dir %s differs from grok session cwd %s (session %s); resume-from-grok-session does not relocate workspace",
				absDir, workspace, id)
		}
		workspace = dirWS
	}
	if workspace == "" {
		// Fallback when summary has no cwd (should be rare after Find).
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		workspace = wd
	}

	// Resolve agent-run session id: --session/--session-id or auto-generate.
	sessionID := strings.TrimSpace(opts.sessionID)
	if sessionID == "" {
		sessionID, err = generateAutoSessionID(opts.prompt, runner, store.Home())
		if err != nil {
			return err
		}
	}
	// Collision on agent-run id (distinct from provider already-mapped).
	if _, err := store.GetSession(sessionID); err == nil {
		return fmt.Errorf("session already exists: %s", sessionID)
	}

	// Empty prompt is allowed for --open / --detach (like normal run). Enforce
	// after P1/P2 validation gates so error leaves keep their specific messages.
	if strings.TrimSpace(opts.prompt) == "" && !opts.openFlag && !opts.detachFlag {
		return fmt.Errorf("prompt is required")
	}

	meta := agentstorage.SessionMeta{
		Runner:                runner,
		SessionID:             sessionID,
		Status:                "running",
		Model:                 opts.model,
		InitialPrompt:         strings.TrimSpace(opts.prompt),
		RunnerSessionID:       id,
		Workspace:             workspace,
		PrependPaths:          append([]string(nil), opts.prependPaths...),
		Env:                   append([]string(nil), opts.envEntries...),
		AgentRunnerConfigHome: strings.TrimSpace(opts.configHome),
	}
	if err := store.CreateSession(sessionID, meta); err != nil {
		return err
	}

	// Existing session with RunnerSessionID → agentui.Run loads it as ResumeSessionID.
	// Open/Detach/KeepTTY match normal run so --detach prints ids and returns early
	// and --open uses the interactive attach path (tests use instant-attach hook).
	// --fork: argv gets grok --fork-session; new agent-run session already created above.
	return agentui.Run(context.Background(), agentui.RunOptions{
		Prompt:                opts.prompt,
		Runner:                runner,
		Model:                 opts.model,
		SessionID:             sessionID,
		AgentRunnerBinary:     opts.agentRunnerBinary,
		AgentRunnerConfigHome: opts.configHome,
		PrependPaths:          opts.prependPaths,
		Env:                   opts.envEntries,
		Color:                 opts.color,
		JSON:                  opts.jsonFlag,
		Workspace:             workspace,
		KeepTerminalAlive:     keepAliveOpenDetach(opts.keepTTY, opts.openFlag, opts.detachFlag),
		Open:                  opts.openFlag,
		Detach:                opts.detachFlag,
		NoSubmit:              opts.noSubmit,
		Fork:                  opts.fork,
		Driver:                mergeHostDriver(agentdriver.Driver{}),
		Store:                 store,
		Stdout:                os.Stdout,
		Stderr:                os.Stderr,
	})
}

type autoSendOrResumeOpts struct {
	jsonFlag                      bool
	model                         string
	modelReasoningEffort          string
	sessionID                     string
	sessionIDFromPrompt           bool
	agentRunner                   string
	agentRunnerBinary             string
	agentRunnerConfigHome         string
	prependPaths                  []string
	envEntries                    []string
	color                         bool
	keepTTY                       bool
	openFlag                      bool
	detachFlag                    bool
	noSubmit                      bool
	dir                           string
	allowRelocateResumeSessionDir bool
	prompt                        string
	defaultRunner                 string
	newTerminal                   bool
	eventBusURL                   string
	eventBusToken                 string
	recorded                      flags.Flags
}

// runAutoSendOrResume classifies a stable --session-id into MODE=run|send|resume
// and dispatches existing run/send/resume semantics via agentrunapi.
func runAutoSendOrResume(opts autoSendOrResumeOpts) error {
	if opts.sessionIDFromPrompt {
		return fmt.Errorf("--auto-send-or-resume and --session-id-from-prompt are mutually exclusive; cannot use both")
	}
	sessionID := strings.TrimSpace(opts.sessionID)
	if sessionID == "" {
		return fmt.Errorf("--auto-send-or-resume requires --session-id")
	}
	if opts.detachFlag && opts.openFlag {
		return fmt.Errorf("--detach and --open are mutually exclusive; cannot use both")
	}
	if opts.detachFlag && opts.jsonFlag {
		return fmt.Errorf("--detach and --json are mutually exclusive; cannot use both")
	}
	if opts.openFlag && opts.jsonFlag {
		return fmt.Errorf("--open and --json are mutually exclusive; cannot use both")
	}

	store, err := openStore()
	if err != nil {
		return err
	}

	runner := resolveCLIRunner(opts.agentRunner, opts.defaultRunner)
	// Shared at-most-once guard for ForceNew NotifyOnOpenPath + library OnTTYStarted.
	var ttyAlreadyNotified bool
	eventBusOpts := EventBusOpts{
		URL:             opts.eventBusURL,
		Token:           opts.eventBusToken,
		WarnWriter:      os.Stderr,
		AlreadyNotified: &ttyAlreadyNotified,
	}
	apiOpts := agentrunapi.Opts{
		SessionID:                     sessionID,
		Prompt:                        opts.prompt,
		WorkspaceDir:                  opts.dir,
		AgentRunner:                   runner,
		AgentRunnerBinary:             opts.agentRunnerBinary,
		RunnerConfigHome:              opts.agentRunnerConfigHome,
		Model:                         opts.model,
		ModelReasoningEffort:          opts.modelReasoningEffort,
		Open:                          opts.openFlag,
		Detach:                        opts.detachFlag,
		NoSubmit:                      opts.noSubmit,
		KeepTTY:                       opts.keepTTY,
		JSON:                          opts.jsonFlag,
		AllowRelocateResumeSessionDir: opts.allowRelocateResumeSessionDir,
		NewTerminal:                   opts.newTerminal,
		Driver:                        mergeHostDriver(agentdriver.Driver{}),
		Env:                           opts.envEntries,
		PrependPaths:                  opts.prependPaths,
		Color:                         opts.color,
		Store:                         store,
		Stdout:                        os.Stdout,
		Stderr:                        os.Stderr,
		// Shared production probe (same as Classify default when Probe is nil).
		Probe: agentrunapi.LifecycleProbe,
		// Wire library TTY lifecycle hooks to HTTP when --event-bus-url is set (nil when empty).
		OnTTYStarted:   WireOnTTYStarted(eventBusOpts),
		OnTTYRestarted: WireOnTTYRestarted(eventBusOpts),
		// Production dispatch keeps full CLI semantics (resume reclaim, new-terminal).
		RunSession: func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta, found bool) error {
			if opts.newTerminal {
				return openAutoInNewTerminal(opts, meta, found, eventBusOpts)
			}
			return autoRunCreate(store, o.SessionID, opts)
		},
		SendLive: func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta) error {
			return autoSendLive(store, meta, opts)
		},
		ResumeSession: func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta) error {
			if opts.newTerminal {
				return openAutoInNewTerminal(opts, meta, true, eventBusOpts)
			}
			return resumeExistingSession(store, meta, resumeRunConfig{
				jsonFlag:                      opts.jsonFlag,
				model:                         opts.model,
				agentRunner:                   opts.agentRunner,
				agentRunnerBinary:             opts.agentRunnerBinary,
				agentRunnerConfigHome:         opts.agentRunnerConfigHome,
				prependPaths:                  opts.prependPaths,
				envEntries:                    opts.envEntries,
				color:                         opts.color,
				keepTTY:                       opts.keepTTY,
				openFlag:                      opts.openFlag,
				detachFlag:                    opts.detachFlag,
				noSubmit:                      opts.noSubmit,
				dir:                           opts.dir,
				allowRelocateResumeSessionDir: opts.allowRelocateResumeSessionDir,
				prompt:                        opts.prompt,
				defaultRunner:                 opts.defaultRunner,
			})
		},
	}
	return agentrunapi.AutoSendOrResume(context.Background(), apiOpts)
}

// openAutoInNewTerminal re-launches the auto-send-or-resume command in a new
// iTerm2 window (ModeForceNew), stripping --new-terminal so the child runs in-process.
// The launcher does not spawn the provider. Follow-up is built via
// agentrunapi.BuildFollowUpCommand (DriverBinary = this executable).
// eventBusOpts shares AlreadyNotified with library OnTTYStarted so ForceNew +
// ModeRun hook publish at most once per open.
func openAutoInNewTerminal(opts autoSendOrResumeOpts, meta agentstorage.SessionMeta, found bool, eventBusOpts EventBusOpts) error {
	exe, err := agentRunExecutable()
	if err != nil {
		return err
	}
	dir, err := resolveNewTerminalDir(opts.dir, meta, found)
	if err != nil {
		return err
	}
	runner := resolveCLIRunner(opts.agentRunner, opts.defaultRunner)
	// Prefer absolute workspace when resolved so child --dir is stable.
	workspaceDir := strings.TrimSpace(opts.dir)
	if workspaceDir == "" && found {
		workspaceDir = strings.TrimSpace(meta.Workspace)
	}
	if absDir := strings.TrimSpace(dir); absDir != "" && workspaceDir != "" {
		// Keep user-facing --dir only when the operator set --dir; otherwise omit
		// so child uses meta.workspace / cwd. When --dir was set, pass resolved dir.
		if strings.TrimSpace(opts.dir) != "" {
			workspaceDir = absDir
		}
	} else if strings.TrimSpace(opts.dir) != "" {
		workspaceDir = dir
	} else {
		workspaceDir = ""
	}

	// Prefer process embedding Driver (spl agent-run); else re-exec this binary alone.
	host := mergeHostDriver(agentdriver.Driver{})
	if strings.TrimSpace(host.Binary) == "" {
		host = agentdriver.Driver{Binary: exe}
	}
	// Carry event-bus flags into the ForceNew child so it can publish if needed.
	eventBusExtra := AppendEventBusFlags(nil, opts.eventBusURL, opts.eventBusToken)
	followUp, err := agentrunapi.BuildFollowUpCommand(agentrunapi.FollowUpOpts{
		Driver:                        host,
		SessionID:                     strings.TrimSpace(opts.sessionID),
		Prompt:                        opts.prompt,
		AgentRunner:                   runner,
		WorkspaceDir:                  workspaceDir,
		NoSubmit:                      opts.noSubmit,
		AllowRelocateResumeSessionDir: opts.allowRelocateResumeSessionDir,
		Open:                          opts.openFlag,
		Detach:                        opts.detachFlag,
		Color:                         opts.color,
		Model:                         opts.model,
		ModelReasoningEffort:          opts.modelReasoningEffort,
		Env:                           append([]string(nil), opts.envEntries...),
		ExtraArgs:                     eventBusExtra,
	})
	if err != nil {
		return err
	}
	// recorded is retained for reconstruct-based debugging / future flag parity;
	// follow-up is intentionally library-built (never includes --new-terminal).
	_ = opts.recorded

	if err := iterm2.OpenConfig(dir, &iterm2.Config{
		Mode:             iterm2.ModeForceNew,
		FollowUpCommands: []string{followUp},
	}); err != nil {
		return err
	}

	// After successful ForceNew open: best-effort agent.tty.started (never fails open).
	// Shares AlreadyNotified with library WireOnTTYStarted installed on apiOpts.
	payloadWorkspace := workspaceDir
	if strings.TrimSpace(payloadWorkspace) == "" {
		payloadWorkspace = dir
	}
	NotifyOnOpenPath("new-terminal", eventBusOpts, strings.TrimSpace(opts.sessionID), runner, payloadWorkspace)
	return nil
}

// agentRunExecutable returns an absolute path to this process binary for re-exec.
func agentRunExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		if len(os.Args) > 0 && os.Args[0] != "" {
			exe = os.Args[0]
		} else {
			return "", fmt.Errorf("resolve agent-run executable: %w", err)
		}
	}
	if abs, absErr := filepath.Abs(exe); absErr == nil {
		exe = abs
	}
	if resolved, evalErr := filepath.EvalSymlinks(exe); evalErr == nil {
		exe = resolved
	}
	return exe, nil
}

// resolveNewTerminalDir picks the workspace for iterm2.OpenConfig:
// --dir if set, else session meta workspace (when found), else process cwd.
func resolveNewTerminalDir(dir string, meta agentstorage.SessionMeta, found bool) (string, error) {
	if resolved, err := resolveRunDir(dir); err != nil {
		return "", err
	} else if resolved != "" {
		return resolved, nil
	}
	if found {
		if w := strings.TrimSpace(meta.Workspace); w != "" {
			abs, err := filepath.Abs(w)
			if err != nil {
				return "", fmt.Errorf("workspace: %w", err)
			}
			if info, err := os.Stat(abs); err == nil && info.IsDir() {
				if real, err := filepath.EvalSymlinks(abs); err == nil {
					return real, nil
				}
				return abs, nil
			}
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cwd: %w", err)
	}
	return cwd, nil
}

func autoSendLive(store agentstorage.Store, meta agentstorage.SessionMeta, opts autoSendOrResumeOpts) error {
	// --open/--detach are accepted but ignored while live: callers (e.g. local-bot)
	// always pass the same CLI for run/send/resume. Live work is enqueue + wait delivery only.
	if opts.openFlag {
		fmt.Fprintln(os.Stderr, "note: --open ignored while session is live; sending follow-up")
	}
	if opts.detachFlag {
		fmt.Fprintln(os.Stderr, "note: --detach ignored while session is live; sending follow-up")
	}
	if opts.newTerminal {
		fmt.Fprintln(os.Stderr, "note: --new-terminal ignored while session is live; sending follow-up")
	}
	if opts.prompt == "" {
		fmt.Fprintln(os.Stderr, "warning: session is live; no message to send")
		return nil
	}

	termID := strings.TrimSpace(meta.TerminalSessionID)
	if termID == "" {
		termID = strings.TrimSpace(meta.SessionID)
	}
	if termID == "" {
		return fmt.Errorf("cannot send: missing terminal_session_id and session_id")
	}

	home := store.Home()
	ttySess, err := agenttty.ResolveByTerminalID(home, termID)
	if err != nil {
		return err
	}
	if !ttySess.TCPReachable {
		return fmt.Errorf("terminal unreachable at %s", ttySess.Registry.ListenAddr)
	}
	provider, ok := agenttty.Get(ttySess.RunnerID)
	if !ok {
		return fmt.Errorf("unknown tty runner: %s", ttySess.RunnerID)
	}

	sess := agentsend.Session{
		Home:              home,
		Runner:            ttySess.RunnerID,
		TerminalSessionID: termID,
		ListenAddr:        ttySess.Registry.ListenAddr,
	}

	enqueuedAt := time.Now()
	id, err := agentsend.EnqueueWith(home, sess, opts.prompt, agentsend.EnqueueOptions{NoSubmit: opts.noSubmit})
	if err != nil {
		return err
	}
	fmt.Println(id)

	waitOpts := agentsend.WaitOptions{
		EnqueuedAt:   enqueuedAt,
		Mode:         agentsend.WaitDefault,
		StartDrainer: true,
	}
	agentsend.StartDrainer(home, sess, provider)
	return agentsend.WaitForDelivery(home, sess, id, waitOpts)
}

func autoRunCreate(store agentstorage.Store, sessionID string, opts autoSendOrResumeOpts) error {
	if opts.detachFlag && opts.openFlag {
		return fmt.Errorf("--detach and --open are mutually exclusive; cannot use both")
	}
	if opts.detachFlag && opts.jsonFlag {
		return fmt.Errorf("--detach and --json are mutually exclusive; cannot use both")
	}
	if opts.prompt == "" && !opts.openFlag && !opts.detachFlag {
		return fmt.Errorf("prompt is required")
	}
	if opts.noSubmit && !opts.openFlag {
		return fmt.Errorf("--no-submit requires --open")
	}
	workspace, err := resolveRunDir(opts.dir)
	if err != nil {
		return err
	}
	runner := resolveCLIRunner(opts.agentRunner, opts.defaultRunner)
	if err := validateRunner(runner); err != nil {
		return err
	}
	if opts.openFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--open requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	if opts.detachFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--detach requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	if err := requireTTYForSessionEnv(runner, opts.prependPaths, opts.envEntries); err != nil {
		return err
	}
	if err := requireTTYForColor(runner, opts.color); err != nil {
		return err
	}
	return agentui.Run(context.Background(), agentui.RunOptions{
		Prompt:                opts.prompt,
		Runner:                runner,
		Model:                 opts.model,
		ModelReasoningEffort:  opts.modelReasoningEffort,
		SessionID:             sessionID,
		AgentRunnerBinary:     opts.agentRunnerBinary,
		AgentRunnerConfigHome: opts.agentRunnerConfigHome,
		PrependPaths:          opts.prependPaths,
		Env:                   opts.envEntries,
		Color:                 opts.color,
		JSON:                  opts.jsonFlag,
		Workspace:             workspace,
		KeepTerminalAlive:     keepAliveOpenDetach(opts.keepTTY, opts.openFlag, opts.detachFlag),
		Open:                  opts.openFlag,
		Detach:                opts.detachFlag,
		NoSubmit:              opts.noSubmit,
		Driver:                mergeHostDriver(agentdriver.Driver{}),
		Store:                 store,
		Stdout:                os.Stdout,
		Stderr:                os.Stderr,
	})
}

// tryResolveSessionMeta looks up a session by bare id or runner/session_id.
// found=false when the session is missing (MODE=run). Ambiguous ids error.
func tryResolveSessionMeta(store agentstorage.Store, ref string) (agentstorage.SessionMeta, bool, error) {
	meta, err := resolveSessionMeta(store, ref)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "session not found") {
			return agentstorage.SessionMeta{}, false, nil
		}
		return agentstorage.SessionMeta{}, false, err
	}
	return meta, true, nil
}

// resolveRunDir validates --dir when set: resolve relative paths against process
// cwd, require exists + directory, prefer EvalSymlinks. Empty means default
// (agentui uses Getwd).
func resolveRunDir(dir string) (string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", nil
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("--dir: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("--dir: path does not exist: %s", abs)
		}
		return "", fmt.Errorf("--dir: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("--dir: %s is not a directory", abs)
	}
	return abs, nil
}
