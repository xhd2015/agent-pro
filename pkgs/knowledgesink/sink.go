package knowledgesink

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	codexsessions "github.com/xhd2015/agent-pro/agent/codex/sessions"
	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

// Mode selects how the propose agent is launched.
type Mode string

const (
	ModeHeadless Mode = "headless" // default: CLI codex/grok via agentui
	ModeOpen     Mode = "open"     // interactive open (TTY)
)

// MessagesFunc injects sessions.Messages for tests (nil → library). Grok tip only.
type MessagesFunc func(grokHome, sessionID string, opts *sessions.MessagesOpts) (*sessions.MessagesResult, error)

// AgentFunc injects the propose agent (nil → agentrunapi.Run / RunJSON).
type AgentFunc func(ctx context.Context, opts agentrunapi.RunOpts, jsonSchema string) (jsonResult string, err error)

// ResolveRunnerFunc resolves Marcus/agent-run session → (runner, runnerSessionID).
type ResolveRunnerFunc func(marcusSessionID string) (runner, runnerSessionID string, err error)

// ResolveSessionDirFunc resolves runner session → absolute session dir (nil → grok/codex libs).
type ResolveSessionDirFunc func(runner, runnerSessionID string) (dir string, err error)

// Opts drives Status and Run.
type Opts struct {
	StateDir      string // required: root that contains knowledge-sink/
	HubDir        string // knowledge-base-hub checkout (required to run agent); caller sets cwd
	SessionID     string // Marcus / agent-run session id
	GrokSessionID string // optional direct runner session id (skips resolve when set)
	Mode          Mode   // empty → headless
	DryRun        bool
	ShowPrompt    bool
	// CreateMR: agent applies hub writes + result.json; host commits/pushes/creates MR.
	CreateMR bool
	// AutoMergeMR implies CreateMR; after MR push, ff-merge and push origin/master.
	AutoMergeMR bool
	// AllowRunning skips the "already sinking" guard. Daemon uses this after it
	// latches manifest status=running for UI before the worker goroutine starts.
	AllowRunning bool
	// Source tags the Marcus trigger for MR title prefixes (auto|ui|slash).
	// Empty → no prefix (bare CLI).
	Source string
	// SkipSessionDirProbe skips Codex/Grok session-file walks when computing
	// Status. For list/auto-pick, a resolved runner session id counts as
	// having content (Codex Find was ~0.3–1s per session via full tree scan).
	SkipSessionDirProbe bool

	// AgentRunner / Model / ModelReasoningEffort are passed to agentrunapi.RunOpts.
	// Empty AgentRunner → library default (grok-tty); callers that want Marcus/
	// shared prefs should resolve before Run.
	AgentRunner          string
	Model                string
	ModelReasoningEffort string
	// Verbose prints concrete agent-run / Codex argv and post-agent ship
	// progress (git commit/push/MR) on Stderr.
	Verbose bool
	// Stderr receives verbose notices (nil → os.Stderr when Verbose).
	Stderr io.Writer

	Driver    agentdriver.Driver
	StoreHome string // agent-run job store; empty → StateDir/knowledge-sink-jobs/<id>

	MessagesFn          MessagesFunc
	AgentFn             AgentFunc
	ResolveFn           ResolveRunnerFunc
	ResolveSessionDirFn ResolveSessionDirFunc
	GitFn               GitRunner
	NowFn               func() time.Time
}

// StatusResult is Status output.
type StatusResult struct {
	OK            bool        `json:"ok"`
	SessionID     string      `json:"session_id,omitempty"`
	GrokSessionID string      `json:"grok_session_id,omitempty"` // runner session id (legacy json name)
	Sink          *StatusView `json:"sink,omitempty"`
	SessionDir    string      `json:"session_dir,omitempty"` // knowledge-sink/<id>
	LastSinkAt    string      `json:"last_sink_at,omitempty"`
	LastPaths     []string    `json:"last_paths,omitempty"`
	LastMRURL     string      `json:"last_mr_url,omitempty"`
	Error         string      `json:"error,omitempty"`
}

// RunResult is Run output (including dry-run / show-prompt).
type RunResult struct {
	OK               bool     `json:"ok"`
	DryRun           bool     `json:"dry_run,omitempty"`
	ShowPrompt       bool     `json:"show_prompt,omitempty"`
	CreateMR         bool     `json:"create_mr,omitempty"`
	AutoMergeMR      bool     `json:"auto_merge_mr,omitempty"`
	Mode             string   `json:"mode,omitempty"`
	SessionID        string   `json:"session_id,omitempty"`
	GrokSessionID    string   `json:"grok_session_id,omitempty"`
	Runner           string   `json:"runner,omitempty"`
	RunnerSessionDir string   `json:"runner_session_dir,omitempty"`
	SinkIndex        int      `json:"sink_index,omitempty"`
	DeltaCount       int      `json:"delta_count,omitempty"`
	ProposalPath     string   `json:"proposal_path,omitempty"`
	ResultJSONPath   string   `json:"result_json_path,omitempty"`
	SessionDir       string   `json:"session_dir,omitempty"` // knowledge-sink/<id>
	HubDir           string   `json:"hub_dir,omitempty"`
	Prompt           string   `json:"prompt,omitempty"`
	HubPaths         []string `json:"hub_paths,omitempty"` // proposed / committed hub paths
	Branch           string   `json:"branch,omitempty"`
	Commit           string   `json:"commit,omitempty"`
	MRURL            string   `json:"mr_url,omitempty"`
	Merged           bool     `json:"merged,omitempty"`
	MergedAt         string   `json:"merged_at,omitempty"`
	Warning          string   `json:"warning,omitempty"`
	LastSinkAt       string   `json:"last_sink_at,omitempty"`
	CursorTimestamp  string   `json:"cursor_timestamp,omitempty"`
	CursorAdvanced   bool     `json:"cursor_advanced,omitempty"`
	HasNewKnowledges *bool    `json:"has_new_knowledges,omitempty"`
	SkipReason       string   `json:"skip_reason,omitempty"`
	Error            string   `json:"error,omitempty"`
}

type agentJSONResult struct {
	OK           bool           `json:"ok"`
	Status       string         `json:"status"`
	SinkIndex    int            `json:"sink_index"`
	ProposalPath string         `json:"proposal_path"`
	Proposals    []proposalItem `json:"proposals"`
	Error        string         `json:"error"`
}

// Status computes sink button state for a session.
func Status(ctx context.Context, opts Opts) (*StatusResult, error) {
	_ = ctx
	if err := validateStateDir(opts); err != nil {
		return nil, err
	}
	key, runner, runnerSID, ok, help, err := resolveIDs(opts)
	if err != nil && strings.TrimSpace(opts.GrokSessionID) == "" && strings.TrimSpace(opts.SessionID) == "" {
		return nil, err
	}
	stateSessionDir := SessionDir(opts.StateDir, storageKey(opts, key))
	manifest, _ := LoadManifest(stateSessionDir)
	if manifest != nil {
		_, _ = ReconcileStaleRunning(stateSessionDir, manifest, nowTime(opts))
		manifest, _ = LoadManifest(stateSessionDir)
	}

	var tip time.Time
	total := 0
	if ok {
		if opts.SkipSessionDirProbe {
			// List/auto-pick: runner session id is enough; avoid Codex Find / message load.
			if strings.TrimSpace(runnerSID) != "" {
				total = 1
			}
		} else if isGrokRunner(runner) || (runner == "" && runnerSID != "") {
			res, merr := fetchMessages(opts, runnerSID)
			if merr != nil {
				return &StatusResult{
					OK:            false,
					SessionID:     key,
					GrokSessionID: runnerSID,
					SessionDir:    stateSessionDir,
					Error:         merr.Error(),
					Sink: &StatusView{
						State:   StateUnavailable,
						Label:   "Sink Knowledge",
						Enabled: false,
						Help:    merr.Error(),
						Error:   merr.Error(),
					},
				}, nil
			}
			if res != nil {
				total = res.Total
				tip = NewestMessageTime(res.Messages)
			}
		} else {
			// Codex (or other): presence of session dir counts as having content.
			if _, derr := resolveRunnerSessionDir(opts, runner, runnerSID); derr == nil {
				total = 1
			} else {
				help = derr.Error()
				ok = false
			}
		}
	}
	out := &StatusResult{
		OK:            true,
		SessionID:     key,
		GrokSessionID: runnerSID,
		SessionDir:    stateSessionDir,
		Sink:          BuildStatus(manifest, tip, total, ok, help),
	}
	if manifest != nil {
		out.LastSinkAt = manifest.LastSinkAt
		out.LastPaths = append([]string(nil), manifest.LastPaths...)
		out.LastMRURL = strings.TrimSpace(manifest.LastMRURL)
	}
	return out, nil
}

// Run performs a sink (propose-only by default, or create-mr / dry-run / show-prompt).
// Propose-only does not advance last_sink_max_message_timestamp; --create-mr does on success.
func Run(ctx context.Context, opts Opts) (*RunResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := validateStateDir(opts); err != nil {
		return nil, err
	}
	mode, err := normalizeMode(opts)
	if err != nil {
		return nil, err
	}
	opts.Mode = mode
	if opts.AutoMergeMR {
		opts.CreateMR = true
	}

	key, runner, runnerSID, ok, help, rerr := resolveIDs(opts)
	if !ok {
		msg := firstNonEmpty(help, errString(rerr), "runner session required")
		return &RunResult{OK: false, SessionID: key, Error: msg}, nil
	}

	st, err := Status(ctx, opts)
	if err != nil {
		return nil, err
	}
	if st == nil || st.Sink == nil {
		return &RunResult{OK: false, SessionID: key, GrokSessionID: runnerSID, Error: "sink status unavailable"}, nil
	}
	if !opts.DryRun && !opts.ShowPrompt && !st.Sink.Enabled && !opts.AllowRunning {
		return &RunResult{
			OK:            false,
			SessionID:     key,
			GrokSessionID: runnerSID,
			Mode:          string(mode),
			Error:         firstNonEmpty(st.Sink.Help, "sink not available"),
		}, nil
	}

	stateSessionDir := SessionDir(opts.StateDir, storageKey(opts, key))
	manifest, _ := LoadManifest(stateSessionDir)
	if manifest == nil {
		manifest = &Manifest{
			Version:         1,
			MarcusSessionID: key,
			NextSinkIndex:   0,
			LastSinkIndex:   -1,
			Status:          statusIdle,
		}
	}
	if !opts.DryRun && !opts.ShowPrompt && strings.EqualFold(manifest.Status, statusRunning) && !opts.AllowRunning {
		return &RunResult{OK: false, SessionID: key, GrokSessionID: runnerSID, Error: "already sinking"}, nil
	}

	runnerSessionDir, derr := resolveRunnerSessionDir(opts, runner, runnerSID)
	if derr != nil {
		return &RunResult{OK: false, SessionID: key, GrokSessionID: runnerSID, Error: "session dir: " + derr.Error()}, nil
	}

	hubDir := strings.TrimSpace(opts.HubDir)
	index := manifest.NextSinkIndex
	prior := buildPriorContext(stateSessionDir, manifest)
	since := strings.TrimSpace(manifest.LastSinkMaxMessageTimestamp)
	proposalAbs := filepath.Join(RunDir(stateSessionDir, index), "proposal.md")
	resultAbs := resultJSONAbsPath(stateSessionDir, index)

	now := nowTime(opts)
	promptIn := PromptInput{
		MarcusSessionID:  key,
		Runner:           firstNonEmpty(runner, "grok"),
		RunnerSessionID:  runnerSID,
		RunnerSessionDir: runnerSessionDir,
		Since:            since,
		SinkIndex:        index,
		ProposalPath:     proposalAbs,
		Prior:            prior,
		CreateMR:         opts.CreateMR,
		ResultJSONPath:   resultAbs,
		BranchDate:       now.Format("2006-01-02"),
	}
	if opts.CreateMR && hubDir != "" {
		if u, uerr := HubGitUser(opts, hubDir); uerr == nil {
			promptIn.GitUser = u
		}
	}

	// Optional delta hint + tip for cursor advance (Grok messages only).
	deltaCount := 0
	var tip time.Time
	if isGrokRunner(runner) || runner == "" {
		if res, merr := fetchMessages(opts, runnerSID); merr == nil && res != nil {
			var after time.Time
			if t, perr := ParseTime(since); perr == nil {
				after = t
			}
			deltaCount = len(FilterMessagesAfter(res.Messages, after))
			tip = NewestMessageTime(res.Messages)
		}
	}

	if opts.ShowPrompt {
		promptIn.ProposalPath = fmt.Sprintf("<would-be sink-%d/proposal.md>", index)
		if opts.CreateMR {
			promptIn.ResultJSONPath = fmt.Sprintf("<would-be sink-%d/result.json>", index)
		}
		return &RunResult{
			OK:               true,
			ShowPrompt:       true,
			CreateMR:         opts.CreateMR,
			AutoMergeMR:      opts.AutoMergeMR,
			Mode:             string(mode),
			SessionID:        key,
			GrokSessionID:    runnerSID,
			Runner:           promptIn.Runner,
			RunnerSessionDir: runnerSessionDir,
			SinkIndex:        index,
			DeltaCount:       deltaCount,
			SessionDir:       stateSessionDir,
			HubDir:           hubDir,
			Prompt:           ShowPromptText(promptIn),
			ProposalPath:     proposalRelPath(index),
			ResultJSONPath:   resultJSONRelPath(index),
			CursorTimestamp:  since,
		}, nil
	}

	if opts.DryRun {
		writeDryRunStages(opts, mode)
		return &RunResult{
			OK:               true,
			DryRun:           true,
			CreateMR:         opts.CreateMR,
			AutoMergeMR:      opts.AutoMergeMR,
			Mode:             string(mode),
			SessionID:        key,
			GrokSessionID:    runnerSID,
			Runner:           promptIn.Runner,
			RunnerSessionDir: runnerSessionDir,
			SinkIndex:        index,
			DeltaCount:       deltaCount,
			SessionDir:       stateSessionDir,
			HubDir:           hubDir,
			ProposalPath:     proposalAbs,
			ResultJSONPath:   resultAbs,
			CursorTimestamp:  since,
		}, nil
	}

	if hubDir == "" {
		return &RunResult{OK: false, SessionID: key, Error: "knowledge-base-hub path empty (--kb-dir)"}, nil
	}
	if fi, err := os.Stat(hubDir); err != nil || !fi.IsDir() {
		return &RunResult{OK: false, SessionID: key, Error: fmt.Sprintf("knowledge-base-hub missing at %s", hubDir)}, nil
	}
	sinkMD := filepath.Join(hubDir, "SINK.md")
	if fi, err := os.Stat(sinkMD); err != nil || fi.IsDir() {
		return &RunResult{OK: false, SessionID: key, Error: "SINK.md not found in hub"}, nil
	}

	if opts.CreateMR {
		// Dirty hub is OK: host ships only git_commit_files from result.json.
		if promptIn.GitUser == "" {
			u, uerr := HubGitUser(opts, hubDir)
			if uerr != nil {
				return &RunResult{OK: false, SessionID: key, GrokSessionID: runnerSID, CreateMR: true, Error: uerr.Error()}, nil
			}
			promptIn.GitUser = u
		}
	}

	manifest.GrokSessionID = runnerSID
	manifest.MarcusSessionID = key
	manifest.Status = statusRunning
	manifest.Error = ""
	manifest.LastPing = FormatTime(now)
	if err := WriteManifest(stateSessionDir, manifest); err != nil {
		return &RunResult{OK: false, SessionID: key, Error: err.Error()}, nil
	}

	fail := func(msg string) *RunResult {
		m, _ := LoadManifest(stateSessionDir)
		if m == nil {
			m = &Manifest{Version: 1, MarcusSessionID: key, LastSinkIndex: -1}
		}
		m.Status = statusFailed
		m.Error = msg
		m.GrokSessionID = runnerSID
		_ = WriteManifest(stateSessionDir, m)
		return &RunResult{
			OK: false, SessionID: key, GrokSessionID: runnerSID, Mode: string(mode),
			CreateMR: opts.CreateMR, AutoMergeMR: opts.AutoMergeMR, Error: msg,
		}
	}

	runDir := RunDir(stateSessionDir, index)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return fail("mkdir sink run: " + err.Error()), nil
	}
	// Reserve sink-N even if the agent fails so retries don't overwrite history.
	manifest.NextSinkIndex = index + 1
	manifest.LastSinkIndex = index
	manifest.GrokSessionID = runnerSID
	manifest.LastPing = FormatTime(nowTime(opts))
	_ = WriteManifest(stateSessionDir, manifest)

	prompt := AgentPrompt(promptIn)
	total := stageTotal(opts)
	step := 0
	step++
	progress(opts, step, total, stageLaunch, launchMsg(opts, mode))
	step++
	waitMsg := "running"
	if opts.CreateMR {
		waitMsg = "wait result.json"
	}
	progress(opts, step, total, stageAgent, waitMsg)
	pingCtx, pingCancel := context.WithCancel(ctx)
	StartPingLoop(pingCtx, stateSessionDir, 0, opts.NowFn)
	jsonOut, agentErr := runAgent(ctx, opts, hubDir, prompt, key, resultAbs)
	pingCancel()
	if agentErr != nil {
		return fail("propose agent: " + agentErr.Error()), nil
	}
	progress(opts, step, total, stageAgent, "ok")

	var proposals []proposalItem
	var proposedHub []string
	if mode == ModeHeadless && strings.TrimSpace(jsonOut) != "" {
		var parsed agentJSONResult
		if jerr := json.Unmarshal([]byte(jsonOut), &parsed); jerr == nil {
			if !parsed.OK || strings.EqualFold(parsed.Status, "failed") {
				return fail(firstNonEmpty(parsed.Error, "propose agent reported failure")), nil
			}
			proposals = parsed.Proposals
			for _, p := range proposals {
				if strings.TrimSpace(p.Path) != "" {
					proposedHub = append(proposedHub, p.Path)
				}
			}
		}
	}

	// Ensure proposal.md exists (agent may have written it; else write from JSON).
	if _, err := os.Stat(proposalAbs); err != nil {
		body := formatProposalMarkdown(promptIn, proposals)
		if werr := os.WriteFile(proposalAbs, []byte(body), 0o644); werr != nil {
			return fail("write proposal.md: " + werr.Error()), nil
		}
	}

	var shipGit *ShipGitResult
	var shipResult *ShipResult
	hubPaths := proposedHub
	if opts.CreateMR {
		step++
		progress(opts, step, total, stageValidate, "result.json")
		verboseNotice(opts, "ship: read/validate %s", resultAbs)
		ship, serr := ReadValidateShipResult(resultAbs, hubDir)
		if serr != nil {
			return fail(serr.Error()), nil
		}
		shipResult = ship
		hubPaths = append([]string(nil), ship.GitCommitFiles.AllPaths()...)
		verboseNotice(opts, "ship: result ok has_new=%v files=%d branch=%s", ship.HasNew(), len(hubPaths), ship.GitBranchName)
		progress(opts, step, total, stageValidate, "ok")

		if ship.HasNew() {
			step++
			progress(opts, step, total, stageShip, "commit+push+MR")
			var shipErr error
			shipGit, shipErr = ShipToMR(opts, hubDir, ship, opts.AutoMergeMR)
			if shipErr != nil {
				// Partial: branch may have pushed before auto-merge failed.
				msg := shipErr.Error()
				if shipGit != nil && shipGit.MRURL != "" {
					return fail(msg + " (mr: " + shipGit.MRURL + ")"), nil
				}
				return fail(msg), nil
			}
			progress(opts, step, total, stageShip, "ok")
			if opts.AutoMergeMR {
				step++
				msg := "ok"
				if shipGit != nil && shipGit.MergedAt != "" {
					msg = "ok @" + shipGit.MergedAt
				}
				progress(opts, step, total, stageMerge, msg)
			}
		} else {
			step++
			progress(opts, step, total, stageShip, "skipped ("+ship.SkipReason+")")
			if opts.AutoMergeMR {
				step++
				progress(opts, step, total, stageMerge, "skipped")
			}
		}
	}

	if latest, _ := LoadManifest(stateSessionDir); latest != nil {
		manifest = latest
	}
	manifest.Status = statusIdle
	manifest.Error = ""
	manifest.Pid = 0
	manifest.GrokSessionID = runnerSID
	// Inconclusive: leave last_sink_at/cursor unchanged so Status stays sinkable
	// ("come back when the session concludes"). Ship / no_new record history.
	inconclusive := opts.CreateMR && shipResult != nil && !shipResult.HasNew() &&
		shipResult.SkipReason == SkipReasonInconclusive
	if !inconclusive {
		manifest.LastSinkAt = FormatTime(now)
	}
	manifest.LastSinkIndex = index
	if manifest.NextSinkIndex <= index {
		manifest.NextSinkIndex = index + 1
	}
	relProposal := proposalRelPath(index)
	manifest.LastPaths = []string{relProposal}
	if opts.CreateMR {
		manifest.LastPaths = append(manifest.LastPaths, resultJSONRelPath(index))
	}
	if len(hubPaths) > 0 {
		manifest.LastHubPaths = append([]string(nil), hubPaths...)
	}
	cursorAdvanced := false
	if opts.CreateMR {
		// Advance cursor when we shipped, or skipped as no_new (done with this slice).
		// Do not advance on inconclusive (come back when the session concludes).
		// Prefer runner tip; if tip is unknown, fall back to last_sink_at so Status
		// does not treat a completed sink as never-sunk.
		advanceCursor := shipResult == nil || shipResult.HasNew() || shipResult.SkipReason == SkipReasonNoNew
		if advanceCursor {
			if !tip.IsZero() {
				manifest.LastSinkMaxMessageTimestamp = FormatTime(tip)
				cursorAdvanced = true
			} else if strings.TrimSpace(manifest.LastSinkMaxMessageTimestamp) == "" {
				manifest.LastSinkMaxMessageTimestamp = manifest.LastSinkAt
				cursorAdvanced = true
			}
		}
		if shipGit != nil {
			manifest.LastBranch = shipGit.Branch
			manifest.LastMRURL = shipGit.MRURL
			manifest.LastCommit = shipGit.Commit
		}
	}
	if err := WriteManifest(stateSessionDir, manifest); err != nil {
		return fail("write manifest: " + err.Error()), nil
	}

	out := &RunResult{
		OK:               true,
		CreateMR:         opts.CreateMR,
		AutoMergeMR:      opts.AutoMergeMR,
		Mode:             string(mode),
		SessionID:        key,
		GrokSessionID:    runnerSID,
		Runner:           promptIn.Runner,
		RunnerSessionDir: runnerSessionDir,
		SinkIndex:        index,
		DeltaCount:       deltaCount,
		ProposalPath:     proposalAbs,
		ResultJSONPath:   resultAbs,
		SessionDir:       stateSessionDir,
		HubDir:           hubDir,
		HubPaths:         hubPaths,
		LastSinkAt:       manifest.LastSinkAt,
		CursorTimestamp:  manifest.LastSinkMaxMessageTimestamp,
		CursorAdvanced:   cursorAdvanced,
	}
	if shipResult != nil && shipResult.HasNewKnowledges != nil {
		out.HasNewKnowledges = BoolPtr(*shipResult.HasNewKnowledges)
		out.SkipReason = shipResult.SkipReason
		if !shipResult.HasNew() {
			switch shipResult.SkipReason {
			case SkipReasonInconclusive:
				out.Warning = "no new knowledges (inconclusive)"
			case SkipReasonNoNew:
				out.Warning = "no new knowledges (no_new)"
			default:
				out.Warning = "no new knowledges"
			}
		}
	}
	if shipGit != nil {
		out.Branch = shipGit.Branch
		out.Commit = shipGit.Commit
		out.MRURL = shipGit.MRURL
		out.Merged = shipGit.Merged
		out.MergedAt = shipGit.MergedAt
		if shipGit.Warning != "" {
			out.Warning = shipGit.Warning
		}
	}
	return out, nil
}

func validateStateDir(opts Opts) error {
	if strings.TrimSpace(opts.StateDir) == "" {
		return fmt.Errorf("state dir is required")
	}
	return nil
}

func normalizeMode(opts Opts) (Mode, error) {
	mode := Mode(strings.TrimSpace(string(opts.Mode)))
	if mode == "" {
		mode = ModeHeadless
	}
	switch mode {
	case ModeHeadless, ModeOpen:
		return mode, nil
	default:
		return "", fmt.Errorf("unknown mode %q (want headless or open)", mode)
	}
}

func storageKey(opts Opts, resolvedMarcus string) string {
	if id := strings.TrimSpace(opts.SessionID); id != "" {
		return id
	}
	if resolvedMarcus != "" {
		return resolvedMarcus
	}
	return strings.TrimSpace(opts.GrokSessionID)
}

func resolveIDs(opts Opts) (marcusID, runner, runnerSessionID string, ok bool, help string, err error) {
	marcusID = strings.TrimSpace(opts.SessionID)
	runnerSessionID = strings.TrimSpace(opts.GrokSessionID)
	if runnerSessionID != "" {
		// Direct id: assume grok unless ResolveFn later overrides via Marcus id.
		return marcusID, "grok", runnerSessionID, true, "", nil
	}
	if marcusID == "" {
		return "", "", "", false, "session id is required (--session or --grok-session)", fmt.Errorf("session id is required")
	}
	r, sid, rerr := resolveRunner(opts, marcusID)
	if rerr != nil {
		return marcusID, "", "", false, rerr.Error(), rerr
	}
	if sid == "" {
		return marcusID, r, "", false, "No runner Session ID on this agent-run session", nil
	}
	if !isGrokRunner(r) && !isCodexRunner(r) {
		return marcusID, r, sid, false, "Runner is " + r + " (grok or codex required)", nil
	}
	return marcusID, r, sid, true, "", nil
}

func resolveRunner(opts Opts, marcusSessionID string) (runner, runnerSessionID string, err error) {
	if opts.ResolveFn != nil {
		return opts.ResolveFn(marcusSessionID)
	}
	store, serr := agentstorage.NewFileStore("")
	if serr != nil || store == nil {
		if serr != nil {
			return "", "", serr
		}
		return "", "", fmt.Errorf("agent-run store unavailable")
	}
	sess, gerr := store.GetSession(marcusSessionID)
	if gerr != nil || sess == nil {
		if gerr != nil {
			return "", "", gerr
		}
		return "", "", fmt.Errorf("agent-run session not found")
	}
	return strings.TrimSpace(sess.Meta.Runner), strings.TrimSpace(sess.Meta.RunnerSessionID), nil
}

func resolveRunnerSessionDir(opts Opts, runner, runnerSessionID string) (string, error) {
	runnerSessionID = strings.TrimSpace(runnerSessionID)
	if runnerSessionID == "" {
		return "", fmt.Errorf("runner session id empty")
	}
	if opts.ResolveSessionDirFn != nil {
		return opts.ResolveSessionDirFn(runner, runnerSessionID)
	}
	if isCodexRunner(runner) {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", herr
		}
		path, err := codexsessions.Find(codexsessions.CodexHomeFromEnv(home), runnerSessionID)
		if err != nil {
			return "", err
		}
		return filepath.Dir(path), nil
	}
	// Default: Grok — Find/Info require an explicit home (empty ≠ ~/.grok).
	grokHome, herr := resolveGrokHome()
	if herr != nil {
		return "", herr
	}
	info, err := sessions.Info(grokHome, runnerSessionID)
	if err != nil {
		return "", err
	}
	if info == nil || strings.TrimSpace(info.SessionDir) == "" {
		return "", fmt.Errorf("grok session dir empty")
	}
	return info.SessionDir, nil
}

// resolveGrokHome returns $GROK_HOME or ~/.grok (Find/Info do not default empty).
func resolveGrokHome() (string, error) {
	if v := strings.TrimSpace(os.Getenv("GROK_HOME")); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".grok"), nil
}

func isGrokRunner(runner string) bool {
	return strings.Contains(strings.ToLower(runner), "grok")
}

func isCodexRunner(runner string) bool {
	return strings.Contains(strings.ToLower(runner), "codex")
}

func fetchMessages(opts Opts, grokSessionID string) (*sessions.MessagesResult, error) {
	mopts := &sessions.MessagesOpts{Limit: 0, LimitSet: true}
	grokHome, err := resolveGrokHome()
	if err != nil {
		return nil, err
	}
	if opts.MessagesFn != nil {
		return opts.MessagesFn(grokHome, grokSessionID, mopts)
	}
	return sessions.Messages(grokHome, grokSessionID, mopts)
}

func runAgent(ctx context.Context, opts Opts, hubDir, prompt, marcusID, resultAbs string) (jsonResult string, err error) {
	mode := opts.Mode
	if mode == "" {
		mode = ModeHeadless
	}
	effective, mappedFrom, err := EffectiveAgentRunner(mode, opts.AgentRunner)
	if err != nil {
		return "", err
	}
	if mappedFrom != "" {
		verboseNotice(opts, "%s runner=%s (mapped from %s)", mode, effective, mappedFrom)
	}
	storeHome := strings.TrimSpace(opts.StoreHome)
	if storeHome == "" {
		storeHome = filepath.Join(opts.StateDir, "knowledge-sink-jobs", SanitizeSessionID(marcusID))
	}
	runOpts := agentrunapi.RunOpts{
		Prompt:               prompt,
		WorkspaceDir:         hubDir,
		StoreHome:            storeHome,
		Driver:               opts.Driver,
		OpenTerminal:         mode == ModeOpen,
		AgentRunner:          effective,
		Model:                strings.TrimSpace(opts.Model),
		ModelReasoningEffort: strings.TrimSpace(opts.ModelReasoningEffort),
		Verbose:              opts.Verbose,
	}
	// Create-MR: wait on sink-N/result.json so --open does not return on TTY idle
	// before the agent finishes writing the ship contract.
	if opts.CreateMR {
		if p := strings.TrimSpace(resultAbs); p != "" {
			runOpts.ResultFile = p
		}
	}
	if opts.AgentFn != nil {
		schema := ""
		if mode == ModeHeadless && !opts.CreateMR {
			schema = ResultSchemaExample
		}
		return opts.AgentFn(ctx, runOpts, schema)
	}
	// Headless: CLI runners via agentui (no detach/PTY). Open keeps agentrunapi TTY path.
	if mode != ModeOpen {
		schema := ""
		if !opts.CreateMR {
			schema = ResultSchemaExample
		}
		return runHeadlessCLI(ctx, opts, runOpts, schema)
	}
	if opts.CreateMR {
		_, err = agentrunapi.Run(ctx, runOpts)
		return "", err
	}
	_, err = agentrunapi.Run(ctx, runOpts)
	return "", err
}

func nowTime(opts Opts) time.Time {
	if opts.NowFn != nil {
		return opts.NowFn()
	}
	return time.Now()
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// ResolveGrokFunc is kept as an alias name for older call sites.
type ResolveGrokFunc = ResolveRunnerFunc
