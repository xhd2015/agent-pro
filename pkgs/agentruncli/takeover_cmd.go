package agentruncli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	codexsessions "github.com/xhd2015/agent-pro/agent/codex/sessions"
	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

const (
	envTakeoverTestHooks = "AGENT_RUN_TAKEOVER_TEST_HOOKS"
)

const takeoverHelp = `
Usage: agent-run takeover [OPTIONS] <session-id>

Adopt a live provider session into agent-run management.

Arguments:
  <session-id>   provider session UUID (Grok/Codex), not agent-run bare id

Options:
  --grok                  alias for --agent-runner=grok-tty
  --codex                 alias for --agent-runner=codex-tty
  --agent-runner RUNNER   explicit runner (grok-tty, codex-tty, ...)
  --color                 force color on the iTerm follow-up TTY child (default)
  --no-color              force no color on the follow-up TTY child (-e NO_COLOR=1)
  --auto-color            leave color to env / parent (no --color, no NO_COLOR)
  --dry-run               report what would be done without acting
  -h, --help              show help
`

// takeoverHooksFile is the L2 injectable snapshot for ListProcs / Lsof / Kill.
type takeoverHooksFile struct {
	Procs     []takeoverHookProc  `json:"procs"`
	OpenFiles map[string][]string `json:"open_files"`
	KillLog   string              `json:"kill_log"`
}

type takeoverHookProc struct {
	PID  int    `json:"pid"`
	PPID int    `json:"ppid"`
	Cmd  string `json:"cmd"`
}

// takeoverLivePID is one live provider process hard-hitting a session.
type takeoverLivePID struct {
	PID  int
	Name string
	Cmd  string
}

// TakeoverDeps injects process/kill/iTerm behavior for tests.
// When nil fields are present, production defaults are used unless hooks env is set.
type TakeoverDeps struct {
	ListProcs func() []procresolve.Proc
	Lsof      func(pid int) []string
	Kill      func(pid int, sig syscall.Signal) error
	WaitDead  func(pid int, timeout time.Duration) bool
	OpenIterm func(dir string, followUp string) error
	// HooksActive is true when AGENT_RUN_TAKEOVER_TEST_HOOKS was loaded (synthetic PIDs).
	HooksActive bool
	// ProcsSnapshot is the full ListProcs result used for ancestry walks.
	ProcsSnapshot []procresolve.Proc
	// KillLog path from hooks file (may be empty).
	KillLog string
}

// takeoverColorMode controls the iTerm follow-up color policy for the TTY child.
type takeoverColorMode int

const (
	// takeoverColorForce is the default: emit run --color on the follow-up.
	takeoverColorForce takeoverColorMode = iota
	// takeoverColorOff forces monochrome via -e NO_COLOR=1 (no --color).
	takeoverColorOff
	// takeoverColorAuto inherits parent/env (no --color, no forced NO_COLOR).
	takeoverColorAuto
)

// runTakeover implements top-level `agent-run takeover`.
// defaultRunner is the value of --agent-runner stripped by Handle (global or
// after the subcommand name — same as run/resume).
func runTakeover(args []string, defaultRunner string) error {
	var grokFlag bool
	var codexFlag bool
	var dryRun bool
	var colorFlag bool
	var noColorFlag bool
	var autoColorFlag bool
	var agentRunner string
	remaining, err := flags.Bool("--grok", &grokFlag).
		Bool("--codex", &codexFlag).
		String("--agent-runner", &agentRunner).
		Bool("--color", &colorFlag).
		Bool("--no-color", &noColorFlag).
		Bool("--auto-color", &autoColorFlag).
		Bool("--dry-run", &dryRun).
		Help("-h,--help", takeoverHelp).
		HelpNoExit().
		Parse(args)
	if err == flags.ErrHelp {
		return nil
	}
	if err != nil {
		return err
	}

	if grokFlag && codexFlag {
		return fmt.Errorf("--grok and --codex are mutually exclusive; cannot use both")
	}

	colorMode, err := resolveTakeoverColorMode(colorFlag, noColorFlag, autoColorFlag)
	if err != nil {
		return err
	}

	// Handle strips --agent-runner into defaultRunner; sub-parse may still see it
	// when called outside Handle. Prefer subcommand value when set.
	explicitRunner := strings.TrimSpace(agentRunner)
	if explicitRunner == "" {
		explicitRunner = strings.TrimSpace(defaultRunner)
	}

	if grokFlag && explicitRunner != "" && explicitRunner != "grok-tty" {
		return fmt.Errorf("--grok conflicts with --agent-runner=%s (expected grok-tty)", explicitRunner)
	}
	if codexFlag && explicitRunner != "" && explicitRunner != "codex-tty" {
		return fmt.Errorf("--codex conflicts with --agent-runner=%s (expected codex-tty)", explicitRunner)
	}

	runner := explicitRunner
	if grokFlag {
		runner = "grok-tty"
	} else if codexFlag {
		runner = "codex-tty"
	}

	if len(remaining) == 0 {
		return fmt.Errorf("takeover requires <session-id>")
	}
	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("takeover requires <session-id>")
	}
	if len(remaining) > 1 {
		return fmt.Errorf("takeover accepts at most one positional <session-id>")
	}

	// No explicit runner: auto-detect via Find under GROK_HOME then CODEX_HOME.
	if strings.TrimSpace(runner) == "" {
		detected, detErr := autoDetectTakeoverRunner(sessionID)
		if detErr != nil {
			return detErr
		}
		runner = detected
	}

	return executeTakeover(sessionID, runner, dryRun, colorMode, nil)
}

// resolveTakeoverColorMode maps CLI flags to a color mode.
// Default when none set: force --color. At most one of the three flags.
func resolveTakeoverColorMode(color, noColor, autoColor bool) (takeoverColorMode, error) {
	n := 0
	if color {
		n++
	}
	if noColor {
		n++
	}
	if autoColor {
		n++
	}
	if n > 1 {
		return 0, fmt.Errorf("--color, --no-color, and --auto-color are mutually exclusive; cannot use more than one")
	}
	switch {
	case noColor:
		return takeoverColorOff, nil
	case autoColor:
		return takeoverColorAuto, nil
	case color:
		return takeoverColorForce, nil
	default:
		// Default: force color so iTerm takeover TTY is not dumb/monochrome.
		return takeoverColorForce, nil
	}
}

// autoDetectTakeoverRunner resolves grok-tty or codex-tty when the operator
// omits --grok/--codex/--agent-runner. Exactly one provider home must contain
// the session id; both → ambiguous; neither → not found.
func autoDetectTakeoverRunner(providerSessionID string) (string, error) {
	providerSessionID = strings.TrimSpace(providerSessionID)
	if providerSessionID == "" {
		return "", fmt.Errorf("takeover requires <session-id>")
	}

	grokHit := false
	if _, err := groksessions.Find(agenttty.GrokHomeForRunner(""), providerSessionID); err == nil {
		grokHit = true
	}
	codexHit := false
	if _, err := codexsessions.Find(agenttty.CodexHomeForRunner(""), providerSessionID); err == nil {
		codexHit = true
	}

	switch {
	case grokHit && codexHit:
		return "", fmt.Errorf("takeover: ambiguous session %s found under both grok and codex; specify --grok or --codex", providerSessionID)
	case grokHit:
		return "grok-tty", nil
	case codexHit:
		return "codex-tty", nil
	default:
		return "", fmt.Errorf("takeover: session %s not found under GROK_HOME or CODEX_HOME (cannot resolve provider)", providerSessionID)
	}
}

// executeTakeover runs the Grok/Codex takeover lifecycle after flag validation.
// deps may be nil (loads env hooks / production defaults).
func executeTakeover(providerSessionID, runner string, dryRun bool, colorMode takeoverColorMode, deps *TakeoverDeps) error {
	providerSessionID = strings.TrimSpace(providerSessionID)
	runner = strings.TrimSpace(runner)
	if runner == "" {
		// Defensive: callers should auto-detect before executeTakeover.
		detected, err := autoDetectTakeoverRunner(providerSessionID)
		if err != nil {
			return err
		}
		runner = detected
	}

	providerKind := ""
	switch runner {
	case "grok", "grok-tty":
		runner = "grok-tty"
		providerKind = "grok"
	case "codex", "codex-tty":
		runner = "codex-tty"
		providerKind = "codex"
	default:
		return fmt.Errorf("takeover: unsupported agent-runner %s", runner)
	}

	if deps == nil {
		d, err := loadTakeoverDepsFromEnv()
		if err != nil {
			return err
		}
		deps = d
	}
	// Snapshot once for LivePIDs + ancestry (production ListProcs is live; hooks are fixed).
	if len(deps.ProcsSnapshot) == 0 && deps.ListProcs != nil {
		deps.ProcsSnapshot = deps.ListProcs()
		snap := deps.ProcsSnapshot
		deps.ListProcs = func() []procresolve.Proc {
			out := make([]procresolve.Proc, len(snap))
			copy(out, snap)
			return out
		}
	}

	workspace, err := findProviderWorkspace(providerKind, providerSessionID)
	if err != nil {
		return err
	}

	// Probe live provider PIDs hard-hitting this session.
	livePIDs := liveProviderPIDsForTakeover(providerKind, providerSessionID, deps.ListProcs, deps.Lsof)

	store, err := openStore()
	if err != nil {
		return err
	}
	home := store.Home()

	// Already managed via live agent-run registry mapping for this provider family.
	if managed, agentID := sessionAlreadyManagedByRegistry(store, home, providerKind, providerSessionID); managed {
		fmt.Fprintf(os.Stderr, "warning: session %s is already managed by agent-run %s; nothing to take over\n",
			providerSessionID, agentID)
		return nil
	}
	// Already managed via process ancestry under agent-run / agent-run-serve.
	if sessionAlreadyManagedByProcess(livePIDs, deps.ProcsSnapshot) {
		fmt.Fprintf(os.Stderr, "warning: session %s is already managed by agent-run; nothing to take over\n",
			providerSessionID)
		return nil
	}

	if workspace != "" {
		if abs, absErr := filepath.Abs(workspace); absErr == nil {
			workspace = abs
		}
	}
	if workspace == "" {
		if wd, wdErr := os.Getwd(); wdErr == nil {
			workspace = wd
		}
	}

	// Resolve existing mapping without creating (needed for dry-run follow-up text).
	mappedMeta, mapped := findMappedProviderMeta(store, providerKind, providerSessionID)
	plannedAgentID := ""
	if mapped {
		plannedAgentID = mappedMeta.SessionID
	} else {
		// Placeholder id for dry-run plan only; real import allocates later.
		plannedAgentID = "new-session"
	}

	followUp, err := buildTakeoverFollowUp(plannedAgentID, runner, workspace, colorMode)
	if err != nil {
		return err
	}

	if dryRun {
		for _, p := range livePIDs {
			name := p.Name
			if name == "" {
				name = providerKind
			}
			fmt.Printf("dry-run: would kill pid %d (%s)\n", p.PID, name)
		}
		fmt.Printf("dry-run: would open iTerm2 with: %s\n", followUp)
		return nil
	}

	// Kill native live PIDs (not under agent-run — already gated above).
	for _, p := range livePIDs {
		if err := killNativePID(deps, p.PID); err != nil {
			return fmt.Errorf("takeover: kill pid %d: %w", p.PID, err)
		}
		fmt.Printf("killed pid %d (%s)\n", p.PID, providerKind)
	}

	// Ensure agent-run session: reuse mapped or CreateSession pre-bind.
	agentSessionID := plannedAgentID
	if mapped {
		agentSessionID = mappedMeta.SessionID
	} else {
		id, createErr := generateAutoSessionID("takeover", runner, home)
		if createErr != nil {
			return createErr
		}
		meta := agentstorage.SessionMeta{
			Runner:          runner,
			SessionID:       id,
			Status:          "finished",
			RunnerSessionID: providerSessionID,
			Workspace:       workspace,
		}
		if err := store.CreateSession(id, meta); err != nil {
			return err
		}
		agentSessionID = id
		followUp, err = buildTakeoverFollowUp(agentSessionID, runner, workspace, colorMode)
		if err != nil {
			return err
		}
	}

	openDir := workspace
	if openDir == "" {
		openDir, _ = os.Getwd()
	}
	if err := deps.OpenIterm(openDir, followUp); err != nil {
		return err
	}

	fmt.Printf("session-id: %s\n", agentSessionID)
	fmt.Printf("provider: %s\n", providerSessionID)
	fmt.Printf("opened new iTerm2 window\n")
	return nil
}

// findProviderWorkspace resolves the provider session under GROK_HOME / CODEX_HOME
// and returns its workspace cwd. Missing session → not-found error.
func findProviderWorkspace(providerKind, providerSessionID string) (string, error) {
	switch providerKind {
	case "grok":
		grokHome := agenttty.GrokHomeForRunner("")
		sess, err := groksessions.Find(grokHome, providerSessionID)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(sess.CWD), nil
	case "codex":
		codexHome := agenttty.CodexHomeForRunner("")
		// Find validates existence; Info fills CWD from session_meta.
		if _, err := codexsessions.Find(codexHome, providerSessionID); err != nil {
			return "", err
		}
		info, err := codexsessions.Info(codexHome, providerSessionID, 1)
		if err != nil {
			// Session exists (Find succeeded); workspace optional.
			return "", nil
		}
		return strings.TrimSpace(info.CWD), nil
	default:
		return "", fmt.Errorf("takeover: unknown provider kind %s", providerKind)
	}
}

func loadTakeoverDepsFromEnv() (*TakeoverDeps, error) {
	deps := &TakeoverDeps{}
	path := strings.TrimSpace(os.Getenv(envTakeoverTestHooks))
	if path == "" {
		deps.ListProcs = procresolve.ListLiveProcs
		deps.Lsof = procresolve.LiveLsof
		deps.Kill = func(pid int, sig syscall.Signal) error {
			return syscall.Kill(pid, sig)
		}
		deps.WaitDead = waitPIDDead
		deps.OpenIterm = openTakeoverIterm
		return deps, nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("takeover: read hooks %s: %w", path, err)
	}
	var doc takeoverHooksFile
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("takeover: parse hooks %s: %w", path, err)
	}

	procs := make([]procresolve.Proc, 0, len(doc.Procs))
	for _, p := range doc.Procs {
		procs = append(procs, procresolve.Proc{PID: p.PID, PPID: p.PPID, Cmd: p.Cmd})
	}
	openFiles := doc.OpenFiles
	if openFiles == nil {
		openFiles = map[string][]string{}
	}
	killLog := strings.TrimSpace(doc.KillLog)

	deps.HooksActive = true
	deps.ProcsSnapshot = procs
	deps.KillLog = killLog
	deps.ListProcs = func() []procresolve.Proc {
		out := make([]procresolve.Proc, len(procs))
		copy(out, procs)
		return out
	}
	deps.Lsof = func(pid int) []string {
		return append([]string(nil), openFiles[strconv.Itoa(pid)]...)
	}
	deps.Kill = func(pid int, sig syscall.Signal) error {
		return appendKillLog(killLog, sig, pid)
	}
	deps.WaitDead = func(pid int, timeout time.Duration) bool {
		// Synthetic hook PIDs: treat as dead after kill is logged.
		_ = pid
		_ = timeout
		return true
	}
	deps.OpenIterm = openTakeoverIterm
	return deps, nil
}

func appendKillLog(path string, sig syscall.Signal, pid int) error {
	if path == "" {
		return nil
	}
	name := signalName(sig)
	line := fmt.Sprintf("%s %d\n", name, pid)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line)
	return err
}

func signalName(sig syscall.Signal) string {
	switch sig {
	case syscall.SIGTERM:
		return "TERM"
	case syscall.SIGKILL:
		return "KILL"
	default:
		return fmt.Sprintf("SIG%d", int(sig))
	}
}

func waitPIDDead(pid int, timeout time.Duration) bool {
	if pid <= 0 {
		return true
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !ttywatch.ProcessAlive(pid) {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return !ttywatch.ProcessAlive(pid)
}

func killNativePID(deps *TakeoverDeps, pid int) error {
	if pid <= 0 {
		return nil
	}
	if err := deps.Kill(pid, syscall.SIGTERM); err != nil && !deps.HooksActive {
		// ESRCH / already dead is OK when not using hooks; continue wait/kill path.
		_ = err
	}
	if deps.WaitDead != nil && deps.WaitDead(pid, 2*time.Second) {
		return nil
	}
	if err := deps.Kill(pid, syscall.SIGKILL); err != nil && !deps.HooksActive {
		if err != syscall.ESRCH {
			return err
		}
	}
	if deps.WaitDead != nil {
		_ = deps.WaitDead(pid, 500*time.Millisecond)
	}
	return nil
}

func openTakeoverIterm(dir, followUp string) error {
	return iterm2.OpenConfig(dir, &iterm2.Config{
		Mode:             iterm2.ModeForceNew,
		FollowUpCommands: []string{followUp},
		SafeInputIgnore:  true,
	})
}

func buildTakeoverFollowUp(agentSessionID, runner, workspace string, colorMode takeoverColorMode) (string, error) {
	agentSessionID = strings.TrimSpace(agentSessionID)
	if agentSessionID == "" {
		return "", fmt.Errorf("takeover: empty agent session id for follow-up")
	}
	exe, err := agentRunExecutable()
	if err != nil {
		// Fall back to bare PATH name so dry-run/plan still works in odd environments.
		exe = "agent-run"
	}
	host := mergeHostDriver(agentdriver.Driver{})
	if strings.TrimSpace(host.Binary) == "" {
		host = agentdriver.Driver{Binary: exe}
	}
	opts := agentrunapi.FollowUpOpts{
		Driver:       host,
		SessionID:    agentSessionID,
		AgentRunner:  runner,
		WorkspaceDir: workspace,
		Open:         true,
	}
	switch colorMode {
	case takeoverColorForce:
		opts.Color = true
	case takeoverColorOff:
		// run has no --no-color; force monochrome on the TTY child via env.
		opts.Color = false
		opts.Env = append(opts.Env, "NO_COLOR=1")
	case takeoverColorAuto:
		// Inherit parent/env: omit --color and do not set NO_COLOR.
		opts.Color = false
	default:
		opts.Color = true
	}
	return agentrunapi.BuildFollowUpCommand(opts)
}

// isTakeoverCodexRunner reports whether meta.runner is a codex family runner.
func isTakeoverCodexRunner(runner string) bool {
	r := strings.TrimSpace(runner)
	return r == "codex" || r == "codex-tty"
}

// metaMatchesProvider reports whether meta.runner is in the same family as providerKind.
func metaMatchesProvider(metaRunner, providerKind string) bool {
	switch providerKind {
	case "grok":
		return isGrokRunner(metaRunner)
	case "codex":
		return isTakeoverCodexRunner(metaRunner)
	default:
		return false
	}
}

func findMappedProviderMeta(store agentstorage.Store, providerKind, providerSessionID string) (agentstorage.SessionMeta, bool) {
	list, err := store.ListSessions()
	if err != nil {
		return agentstorage.SessionMeta{}, false
	}
	providerSessionID = strings.TrimSpace(providerSessionID)
	for _, m := range list {
		if !metaMatchesProvider(m.Runner, providerKind) {
			continue
		}
		if strings.TrimSpace(m.RunnerSessionID) == providerSessionID {
			return m, true
		}
	}
	return agentstorage.SessionMeta{}, false
}

// sessionAlreadyManagedByRegistry reports whether any agent-run meta maps the
// provider UUID (same runner family) and has a live registry PID.
func sessionAlreadyManagedByRegistry(store agentstorage.Store, home, providerKind, providerSessionID string) (bool, string) {
	list, err := store.ListSessions()
	if err != nil {
		return false, ""
	}
	providerSessionID = strings.TrimSpace(providerSessionID)
	for _, m := range list {
		if !metaMatchesProvider(m.Runner, providerKind) {
			continue
		}
		if strings.TrimSpace(m.RunnerSessionID) != providerSessionID {
			continue
		}
		termID := strings.TrimSpace(m.TerminalSessionID)
		if termID == "" {
			termID = strings.TrimSpace(m.SessionID)
		}
		if termID == "" {
			continue
		}
		ttySess, err := agenttty.ResolveByTerminalID(home, termID)
		if err != nil {
			if termID != m.SessionID {
				ttySess, err = agenttty.ResolveByTerminalID(home, m.SessionID)
			}
		}
		if err != nil {
			continue
		}
		pid := ttySess.Registry.PID
		if pid > 0 && ttywatch.ProcessAlive(pid) {
			return true, m.SessionID
		}
	}
	return false, ""
}

// liveProviderPIDsForTakeover finds provider runners whose open files hard-hit sessionID.
// providerKind is "grok" or "codex". Accepts production path markers and UUID path segments.
func liveProviderPIDsForTakeover(providerKind, sessionID string, listProcs func() []procresolve.Proc, lsof func(int) []string) []takeoverLivePID {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}
	if listProcs == nil {
		listProcs = procresolve.ListLiveProcs
	}
	if lsof == nil {
		lsof = procresolve.LiveLsof
	}
	want := strings.ToLower(sessionID)
	var hits []takeoverLivePID
	for _, p := range listProcs() {
		if !cmdMatchesProviderRunner(p.Cmd, providerKind) {
			continue
		}
		if !openFilesHitProviderSession(lsof(p.PID), providerKind, want) {
			continue
		}
		name := ""
		if fields := strings.Fields(p.Cmd); len(fields) > 0 {
			name = filepath.Base(fields[0])
		}
		hits = append(hits, takeoverLivePID{
			PID:  p.PID,
			Name: name,
			Cmd:  p.Cmd,
		})
	}
	sort.Slice(hits, func(i, j int) bool {
		return hits[i].PID < hits[j].PID
	})
	return hits
}

func cmdMatchesProviderRunner(cmd, providerKind string) bool {
	switch providerKind {
	case "grok":
		return procresolve.IsGrokRunner(cmd)
	case "codex":
		return takeoverClassifyCmd(cmd) == "codex"
	default:
		return false
	}
}

func openFilesHitProviderSession(paths []string, providerKind, wantLower string) bool {
	if wantLower == "" {
		return false
	}
	for _, p := range paths {
		if kind, id, ok := procresolve.ParseSessionFromPath(p); ok && kind == providerKind && strings.EqualFold(id, wantLower) {
			return true
		}
		// Fallback: UUID appears as a path segment (or inside rollout filename).
		slash := filepath.ToSlash(p)
		for _, seg := range strings.Split(slash, "/") {
			if strings.EqualFold(seg, wantLower) {
				return true
			}
			// Codex rollouts: rollout-…-<uuid>.jsonl
			if strings.Contains(strings.ToLower(seg), wantLower) {
				return true
			}
		}
	}
	return false
}

// sessionAlreadyManagedByProcess walks PPID chains of live session PIDs in the
// provided process snapshot; any agent-run / agent-run-serve ancestor means managed.
func sessionAlreadyManagedByProcess(livePIDs []takeoverLivePID, procs []procresolve.Proc) bool {
	if len(livePIDs) == 0 || len(procs) == 0 {
		return false
	}
	byPID := make(map[int]procresolve.Proc, len(procs))
	for _, p := range procs {
		byPID[p.PID] = p
	}
	for _, live := range livePIDs {
		seen := map[int]bool{}
		pid := live.PID
		for pid > 0 && !seen[pid] {
			seen[pid] = true
			p, ok := byPID[pid]
			if !ok {
				break
			}
			if role := takeoverClassifyCmd(p.Cmd); role == "agent-run" || role == "agent-run-serve" {
				return true
			}
			pid = p.PPID
		}
	}
	return false
}

// takeoverClassifyCmd mirrors procresolve role classification for ancestry walks
// (procresolve.classifyCmd is unexported).
func takeoverClassifyCmd(cmd string) string {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return "other"
	}
	base := filepath.Base(fields[0])
	switch base {
	case "agent-run":
		for _, f := range fields[1:] {
			if f == "serve" {
				return "agent-run-serve"
			}
			if strings.HasPrefix(f, "-") {
				break
			}
		}
		return "agent-run"
	case "grok":
		return "grok"
	case "codex":
		return "codex"
	default:
		return "other"
	}
}
