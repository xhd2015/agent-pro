package knowledgesink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func TestResolveGrokHomeDefaults(t *testing.T) {
	t.Setenv("GROK_HOME", "")
	got, err := resolveGrokHome()
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".grok")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	t.Setenv("GROK_HOME", "/tmp/custom-grok-home")
	got, err = resolveGrokHome()
	if err != nil || got != "/tmp/custom-grok-home" {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestResolveRunnerSessionDirUsesGrokHome(t *testing.T) {
	// Real session under ~/.grok — proves empty-home bug is gone.
	const id = "01a02eea-f452-7c00-9a58-11ca13866d5f"
	dir, err := resolveRunnerSessionDir(Opts{}, "grok-tty", id)
	if err != nil {
		t.Skipf("session %s not on this machine: %v", id, err)
	}
	if !strings.Contains(dir, id) {
		t.Fatalf("dir=%q", dir)
	}
}

func TestFormatParseTime(t *testing.T) {
	now := time.Date(2026, 3, 20, 20, 15, 3, 0, time.FixedZone("CST", 8*3600))
	s := FormatTime(now)
	if !strings.Contains(s, "2026-03-20 20:15:03") {
		t.Fatalf("format = %q", s)
	}
	parsed, err := ParseTime(s)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Unix() != now.Unix() {
		t.Fatalf("roundtrip %v vs %v", parsed, now)
	}
}

func TestTipAfterMax(t *testing.T) {
	a := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	b := a.Add(time.Hour)
	if TipAfterMax(a, b) {
		t.Fatal("tip before max must be caught up")
	}
	if TipAfterMax(b, b) {
		t.Fatal("equal tip must be caught up")
	}
	if !TipAfterMax(b, a) {
		t.Fatal("tip after max must be behind")
	}
	// Unknown tip + existing cursor → not after (sunk).
	if TipAfterMax(time.Time{}, a) {
		t.Fatal("zero tip with cursor must not claim behind")
	}
	if !TipAfterMax(time.Time{}, time.Time{}) {
		t.Fatal("zero tip and zero cursor: never-sunk path uses neverSunk before this")
	}
}

func TestBuildStatusStates(t *testing.T) {
	tip := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	ready := BuildStatus(nil, tip, 3, true, "")
	if ready.State != StateReady || !ready.Enabled || ready.Label != "Sink Knowledge" {
		t.Fatalf("ready = %+v", ready)
	}
	unavail := BuildStatus(nil, tip, 3, false, "needs runner")
	if unavail.State != StateUnavailable || unavail.Enabled {
		t.Fatalf("unavailable = %+v", unavail)
	}
	m := &Manifest{
		LastSinkMaxMessageTimestamp: FormatTime(tip),
		Status:                      statusIdle,
	}
	sunk := BuildStatus(m, tip, 3, true, "")
	if sunk.State != StateSunk || sunk.Enabled || sunk.Label != "Sinked" {
		t.Fatalf("sunk = %+v", sunk)
	}
}

func TestBuildStatusHistoryWithoutCursor(t *testing.T) {
	sinkAt := time.Date(2026, 8, 25, 12, 20, 2, 0, time.FixedZone("CST", 8*3600))
	m := &Manifest{
		LastSinkAt: FormatTime(sinkAt),
		Status:     statusIdle,
	}
	// Tip unknown → sunk (not ready/never-sunk); auto-pick must skip.
	sunk := BuildStatus(m, time.Time{}, 3, true, "")
	if sunk.State != StateSunk || sunk.Enabled || IsAutoSinkable(sunk) {
		t.Fatalf("history+no cursor+zero tip: want sunk, got %+v", sunk)
	}
	if AutoSinkWhy(sunk) == "never sunk" {
		t.Fatalf("why must not be never sunk: %q", AutoSinkWhy(sunk))
	}
	// Tip before / equal last_sink_at → still sunk.
	before := BuildStatus(m, sinkAt.Add(-time.Hour), 3, true, "")
	if before.State != StateSunk || IsAutoSinkable(before) {
		t.Fatalf("tip before last_sink_at: want sunk, got %+v", before)
	}
	// Tip after last_sink_at → behind / sinkable (real new work).
	behind := BuildStatus(m, sinkAt.Add(time.Hour), 3, true, "")
	if behind.State != StateBehind || !behind.Enabled || !IsAutoSinkable(behind) {
		t.Fatalf("tip after last_sink_at: want behind, got %+v", behind)
	}
	if AutoSinkWhy(behind) == "never sunk" {
		t.Fatalf("behind why must not be never sunk: %q", AutoSinkWhy(behind))
	}
	// No history and no cursor → still ready / never sunk.
	fresh := BuildStatus(&Manifest{Status: statusIdle}, sinkAt, 3, true, "")
	if fresh.State != StateReady || !IsAutoSinkable(fresh) || AutoSinkWhy(fresh) != "never sunk" {
		t.Fatalf("no history: want ready/never sunk, got %+v why=%q", fresh, AutoSinkWhy(fresh))
	}
}

func TestStatusSkipSessionDirProbe(t *testing.T) {
	state := t.TempDir()
	dirProbe := 0
	res, err := Status(context.Background(), Opts{
		StateDir:  state,
		SessionID: "marcus-1",
		ResolveFn: func(string) (string, string, error) {
			return "codex-tty", "runner-1", nil
		},
		ResolveSessionDirFn: func(string, string) (string, error) {
			dirProbe++
			return "", fmt.Errorf("should not probe")
		},
		SkipSessionDirProbe: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if dirProbe != 0 {
		t.Fatalf("dirProbe=%d", dirProbe)
	}
	if res == nil || res.Sink == nil || res.Sink.State != StateReady || !res.Sink.Enabled {
		t.Fatalf("res=%+v sink=%+v", res, res.Sink)
	}
}

func TestAgentPromptContainsSINKAndPrior(t *testing.T) {
	p := AgentPrompt(PromptInput{
		MarcusSessionID:  "m1",
		Runner:           "grok-tty",
		RunnerSessionID:  "g1",
		RunnerSessionDir: "/tmp/sess",
		Since:            "2026-01-01 00:00:00 +0000 UTC",
		SinkIndex:        1,
		ProposalPath:     "/tmp/state/sink-1/proposal.md",
		Prior: PriorSinkContext{
			HasPrior:     true,
			LastSinkAt:   "yesterday",
			LastHubPaths: []string{"topics/a.md"},
			PriorRunDirs: []string{"/tmp/state/sink-0"},
		},
	})
	for _, want := range []string{
		"./SINK.md",
		"session_dir: /tmp/sess",
		"Do NOT write",
		"Conclusion gate",
		"Novelty gate",
		"last_hub_paths",
		"topics/a.md",
		"incremental",
		"/tmp/state/sink-1/proposal.md",
	} {
		if !strings.Contains(p, want) {
			t.Fatalf("prompt missing %q:\n%s", want, p)
		}
	}
	if strings.Contains(p, "MAINTENANCE.md") || strings.Contains(p, "__meta_knowledge__") {
		t.Fatal("prompt must not mention MAINTENANCE/meta as secondary")
	}
	if strings.Contains(p, "## Output") {
		t.Fatal("propose-only prompt must not include Output section")
	}
}

func TestRun_DryRunAndShowPromptNoWrite(t *testing.T) {
	state := t.TempDir()
	hub := filepath.Join(t.TempDir(), "knowledge-base-hub")
	if err := os.MkdirAll(hub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "SINK.md"), []byte("# sink\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runnerDir := t.TempDir()
	t0 := time.Date(2026, 3, 20, 10, 0, 0, 0, time.Local)
	opts := Opts{
		StateDir:            state,
		HubDir:              hub,
		SessionID:           "marcus-1",
		ResolveFn:           func(string) (string, string, error) { return "grok-tty", "grok-1", nil },
		ResolveSessionDirFn: func(string, string) (string, error) { return runnerDir, nil },
		MessagesFn: func(string, string, *sessions.MessagesOpts) (*sessions.MessagesResult, error) {
			return &sessions.MessagesResult{
				Total: 1,
				Messages: []sessions.ChatMessage{
					{Kind: sessions.MessageKindUser, Text: "hello", Timestamp: t0},
				},
			}, nil
		},
		AgentFn: func(context.Context, agentrunapi.RunOpts, string) (string, error) {
			t.Fatal("agent must not run")
			return "", nil
		},
	}

	opts.DryRun = true
	dry, err := Run(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !dry.OK || !dry.DryRun {
		t.Fatalf("dry = %+v", dry)
	}
	if _, err := os.Stat(SessionDir(state, "marcus-1")); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not create session dir, err=%v", err)
	}

	opts.DryRun = false
	opts.ShowPrompt = true
	sp, err := Run(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !sp.OK || !sp.ShowPrompt || !strings.Contains(sp.Prompt, "SINK.md") {
		t.Fatalf("show-prompt = %+v", sp)
	}
	if !strings.Contains(sp.Prompt, runnerDir) {
		t.Fatalf("prompt missing session dir: %s", sp.Prompt)
	}
}

func TestRun_ProposeOnlyWritesProposalDoesNotAdvanceCursor(t *testing.T) {
	state := t.TempDir()
	hub := filepath.Join(t.TempDir(), "knowledge-base-hub")
	if err := os.MkdirAll(hub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "SINK.md"), []byte("# sink\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runnerDir := t.TempDir()
	t0 := time.Date(2026, 3, 20, 10, 0, 0, 0, time.Local)
	t1 := t0.Add(5 * time.Minute)
	cursorBefore := FormatTime(t0)
	// Seed a prior cursor so we can assert it does not advance.
	sessDir := SessionDir(state, "marcus-1")
	if err := WriteManifest(sessDir, &Manifest{
		Version:                     1,
		MarcusSessionID:             "marcus-1",
		LastSinkMaxMessageTimestamp: cursorBefore,
		NextSinkIndex:               0,
		LastSinkIndex:               -1,
		Status:                      statusIdle,
	}); err != nil {
		t.Fatal(err)
	}

	opts := Opts{
		StateDir:            state,
		HubDir:              hub,
		SessionID:           "marcus-1",
		Mode:                ModeHeadless,
		NowFn:               func() time.Time { return t1.Add(time.Minute) },
		ResolveFn:           func(string) (string, string, error) { return "grok-tty", "grok-1", nil },
		ResolveSessionDirFn: func(string, string) (string, error) { return runnerDir, nil },
		MessagesFn: func(string, string, *sessions.MessagesOpts) (*sessions.MessagesResult, error) {
			return &sessions.MessagesResult{
				Total: 2,
				Messages: []sessions.ChatMessage{
					{Kind: sessions.MessageKindUser, Text: "hello", Timestamp: t0},
					{Kind: sessions.MessageKindResponse, Text: "world", Timestamp: t1},
				},
			}, nil
		},
		AgentFn: func(_ context.Context, o agentrunapi.RunOpts, schema string) (string, error) {
			if schema == "" {
				t.Fatal("headless must pass JSON schema")
			}
			if o.WorkspaceDir != hub {
				t.Fatalf("cwd/workspace = %q want hub", o.WorkspaceDir)
			}
			if !strings.Contains(o.Prompt, "SINK.md") || !strings.Contains(o.Prompt, runnerDir) {
				t.Fatalf("prompt = %q", o.Prompt)
			}
			if strings.Contains(o.Prompt, "MAINTENANCE.md") {
				t.Fatal("unexpected MAINTENANCE.md in prompt")
			}
			raw, _ := json.Marshal(agentJSONResult{
				OK:     true,
				Status: "proposed",
				Proposals: []proposalItem{
					{Path: "topics/x.md", Kind: "new", Rationale: "r", Evidence: "e"},
				},
			})
			return string(raw), nil
		},
	}

	res, err := Run(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.SinkIndex != 0 {
		t.Fatalf("run = %+v", res)
	}
	if _, err := os.Stat(res.ProposalPath); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(res.ProposalPath)
	if !strings.Contains(string(body), "topics/x.md") {
		t.Fatalf("proposal.md = %s", body)
	}
	// Hub must not gain topic files from this library path.
	if _, err := os.Stat(filepath.Join(hub, "topics", "x.md")); !os.IsNotExist(err) {
		t.Fatal("propose-only must not write hub leaves")
	}

	man, err := LoadManifest(sessDir)
	if err != nil || man == nil {
		t.Fatalf("manifest: %v %+v", err, man)
	}
	if man.LastSinkMaxMessageTimestamp != cursorBefore {
		t.Fatalf("cursor advanced: %q → %q", cursorBefore, man.LastSinkMaxMessageTimestamp)
	}
	if man.LastPaths[0] != "sink-0/proposal.md" {
		t.Fatalf("last_paths = %v", man.LastPaths)
	}
	if len(man.LastHubPaths) != 1 || man.LastHubPaths[0] != "topics/x.md" {
		t.Fatalf("last_hub_paths = %v", man.LastHubPaths)
	}
}

func TestRun_RequiresSINKMd(t *testing.T) {
	state := t.TempDir()
	hub := t.TempDir() // no SINK.md
	opts := Opts{
		StateDir:            state,
		HubDir:              hub,
		SessionID:           "marcus-1",
		ResolveFn:           func(string) (string, string, error) { return "grok-tty", "grok-1", nil },
		ResolveSessionDirFn: func(string, string) (string, error) { return t.TempDir(), nil },
		MessagesFn: func(string, string, *sessions.MessagesOpts) (*sessions.MessagesResult, error) {
			return &sessions.MessagesResult{
				Total: 1,
				Messages: []sessions.ChatMessage{
					{Kind: sessions.MessageKindUser, Text: "x", Timestamp: time.Now()},
				},
			}, nil
		},
		AgentFn: func(context.Context, agentrunapi.RunOpts, string) (string, error) {
			t.Fatal("must not run")
			return "", nil
		},
	}
	res, err := Run(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if res.OK || !strings.Contains(res.Error, "SINK.md") {
		t.Fatalf("want SINK.md error, got %+v", res)
	}
}

func TestRun_PassesAgentRunnerModelAndReasoning(t *testing.T) {
	state := t.TempDir()
	hub := filepath.Join(t.TempDir(), "knowledge-base-hub")
	if err := os.MkdirAll(hub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "SINK.md"), []byte("# sink\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runnerDir := t.TempDir()
	t0 := time.Now()
	var got agentrunapi.RunOpts
	opts := Opts{
		StateDir:             state,
		HubDir:               hub,
		SessionID:            "marcus-prefs",
		Mode:                 ModeOpen,
		AgentRunner:          "codex-tty",
		Model:                "gpt-5.6-luna",
		ModelReasoningEffort: "max",
		Verbose:              true,
		ResolveFn:            func(string) (string, string, error) { return "codex-tty", "c1", nil },
		ResolveSessionDirFn:  func(string, string) (string, error) { return runnerDir, nil },
		MessagesFn: func(string, string, *sessions.MessagesOpts) (*sessions.MessagesResult, error) {
			return &sessions.MessagesResult{
				Total: 1,
				Messages: []sessions.ChatMessage{
					{Kind: sessions.MessageKindUser, Text: "hi", Timestamp: t0},
				},
			}, nil
		},
		AgentFn: func(_ context.Context, o agentrunapi.RunOpts, _ string) (string, error) {
			got = o
			return "", nil
		},
	}
	res, err := Run(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("run = %+v", res)
	}
	if !got.OpenTerminal {
		t.Fatal("open mode must set OpenTerminal")
	}
	if got.ResultFile != "" {
		t.Fatalf("propose-only open must not set ResultFile, got %q", got.ResultFile)
	}
	if got.AgentRunner != "codex-tty" || got.Model != "gpt-5.6-luna" || got.ModelReasoningEffort != "max" {
		t.Fatalf("RunOpts prefs = %+v", got)
	}
	if !got.Verbose {
		t.Fatal("Verbose must plumb to RunOpts")
	}
}

func TestRun_HeadlessMapsTTYRunnerToCLI(t *testing.T) {
	state := t.TempDir()
	hub := filepath.Join(t.TempDir(), "knowledge-base-hub")
	if err := os.MkdirAll(hub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "SINK.md"), []byte("# sink\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runnerDir := t.TempDir()
	t0 := time.Now()
	var got agentrunapi.RunOpts
	var stderr bytes.Buffer
	opts := Opts{
		StateDir:             state,
		HubDir:               hub,
		SessionID:            "marcus-headless-map",
		Mode:                 ModeHeadless,
		AgentRunner:          "codex-tty",
		Model:                "gpt-5.6-luna",
		ModelReasoningEffort: "max",
		Verbose:              true,
		Stderr:               &stderr,
		ResolveFn:            func(string) (string, string, error) { return "codex-tty", "c1", nil },
		ResolveSessionDirFn:  func(string, string) (string, error) { return runnerDir, nil },
		MessagesFn: func(string, string, *sessions.MessagesOpts) (*sessions.MessagesResult, error) {
			return &sessions.MessagesResult{
				Total: 1,
				Messages: []sessions.ChatMessage{
					{Kind: sessions.MessageKindUser, Text: "hi", Timestamp: t0},
				},
			}, nil
		},
		AgentFn: func(_ context.Context, o agentrunapi.RunOpts, schema string) (string, error) {
			got = o
			if schema == "" {
				t.Fatal("headless propose must pass JSON schema")
			}
			raw, _ := json.Marshal(agentJSONResult{OK: true, Status: "ok"})
			return string(raw), nil
		},
	}
	res, err := Run(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("run = %+v", res)
	}
	if got.OpenTerminal {
		t.Fatal("headless must not open terminal")
	}
	if got.AgentRunner != "codex" {
		t.Fatalf("effective runner = %q want codex", got.AgentRunner)
	}
	if !strings.Contains(stderr.String(), "mapped from codex-tty") {
		t.Fatalf("stderr missing map notice:\n%s", stderr.String())
	}
}

func TestRun_CreateMROpenWaitsOnResultFile(t *testing.T) {
	state := t.TempDir()
	hub, _ := setupHubRemote(t)
	runnerDir := t.TempDir()
	t0 := time.Now()
	var got agentrunapi.RunOpts
	opts := Opts{
		StateDir:            state,
		HubDir:              hub,
		SessionID:           "marcus-open-mr",
		Mode:                ModeOpen,
		CreateMR:            true,
		ResolveFn:           func(string) (string, string, error) { return "codex-tty", "c1", nil },
		ResolveSessionDirFn: func(string, string) (string, error) { return runnerDir, nil },
		MessagesFn: func(string, string, *sessions.MessagesOpts) (*sessions.MessagesResult, error) {
			return &sessions.MessagesResult{
				Total: 1,
				Messages: []sessions.ChatMessage{
					{Kind: sessions.MessageKindUser, Text: "hi", Timestamp: t0},
				},
			}, nil
		},
		AgentFn: func(_ context.Context, o agentrunapi.RunOpts, schema string) (string, error) {
			got = o
			if schema != "" {
				t.Fatal("create-mr open must not pass propose schema")
			}
			want := resultJSONAbsPath(SessionDir(state, "marcus-open-mr"), 0)
			if o.ResultFile != want {
				t.Fatalf("ResultFile = %q want %q", o.ResultFile, want)
			}
			if !o.OpenTerminal {
				t.Fatal("expected OpenTerminal")
			}
			if err := os.WriteFile(filepath.Join(hub, "topics-open.md"), []byte("x\n"), 0o644); err != nil {
				return "", err
			}
			body, _ := json.Marshal(ShipResult{
				HasNewKnowledges: BoolPtr(true),
				GitCommitMsg:     "docs(kb): open mr",
				GitBranchName:    "tester/2026-03-24-open-mr",
				GitCommitFiles: ShipCommitFiles{
					Add: []string{"topics-open.md"},
				},
			})
			if err := os.WriteFile(want, body, 0o644); err != nil {
				return "", err
			}
			return "", nil
		},
	}
	res, err := Run(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || !res.CreateMR {
		t.Fatalf("run = %+v", res)
	}
	if got.ResultFile == "" {
		t.Fatal("ResultFile empty")
	}
}

func TestRun_CreateMRSkipNoNewAdvancesCursor(t *testing.T) {
	state := t.TempDir()
	hub, _ := setupHubRemote(t)
	runnerDir := t.TempDir()
	t0 := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	opts := Opts{
		StateDir:            state,
		HubDir:              hub,
		SessionID:           "marcus-skip-nonew",
		Mode:                ModeHeadless,
		CreateMR:            true,
		AutoMergeMR:         true,
		// Tip/cursor advance is computed from Grok messages only.
		ResolveFn:           func(string) (string, string, error) { return "grok-tty", "g1", nil },
		ResolveSessionDirFn: func(string, string) (string, error) { return runnerDir, nil },
		MessagesFn: func(string, string, *sessions.MessagesOpts) (*sessions.MessagesResult, error) {
			return &sessions.MessagesResult{
				Total: 1,
				Messages: []sessions.ChatMessage{
					{Kind: sessions.MessageKindUser, Text: "hi", Timestamp: t0},
				},
			}, nil
		},
		AgentFn: func(_ context.Context, o agentrunapi.RunOpts, _ string) (string, error) {
			body, _ := json.Marshal(ShipResult{
				HasNewKnowledges: BoolPtr(false),
				SkipReason:       SkipReasonNoNew,
			})
			return "", os.WriteFile(o.ResultFile, body, 0o644)
		},
		NowFn: func() time.Time { return t0.Add(time.Hour) },
	}
	res, err := Run(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.MRURL != "" {
		t.Fatalf("run = %+v", res)
	}
	if res.HasNewKnowledges == nil || *res.HasNewKnowledges {
		t.Fatalf("has_new = %+v", res.HasNewKnowledges)
	}
	if res.SkipReason != SkipReasonNoNew {
		t.Fatalf("skip=%q", res.SkipReason)
	}
	if !res.CursorAdvanced {
		t.Fatal("expected cursor advanced for no_new")
	}
	m, err := LoadManifest(SessionDir(state, "marcus-skip-nonew"))
	if err != nil || m == nil || strings.TrimSpace(m.LastSinkMaxMessageTimestamp) == "" {
		t.Fatalf("manifest cursor = %+v err=%v", m, err)
	}
}

func TestRun_CreateMRSkipInconclusiveKeepsCursor(t *testing.T) {
	state := t.TempDir()
	hub, _ := setupHubRemote(t)
	runnerDir := t.TempDir()
	t0 := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	opts := Opts{
		StateDir:            state,
		HubDir:              hub,
		SessionID:           "marcus-skip-inconclusive",
		Mode:                ModeHeadless,
		CreateMR:            true,
		AutoMergeMR:         true,
		ResolveFn:           func(string) (string, string, error) { return "codex-tty", "c1", nil },
		ResolveSessionDirFn: func(string, string) (string, error) { return runnerDir, nil },
		MessagesFn: func(string, string, *sessions.MessagesOpts) (*sessions.MessagesResult, error) {
			return &sessions.MessagesResult{
				Total: 1,
				Messages: []sessions.ChatMessage{
					{Kind: sessions.MessageKindUser, Text: "hi", Timestamp: t0},
				},
			}, nil
		},
		AgentFn: func(_ context.Context, o agentrunapi.RunOpts, _ string) (string, error) {
			body, _ := json.Marshal(ShipResult{
				HasNewKnowledges: BoolPtr(false),
				SkipReason:       SkipReasonInconclusive,
			})
			return "", os.WriteFile(o.ResultFile, body, 0o644)
		},
		NowFn: func() time.Time { return t0.Add(time.Hour) },
	}
	res, err := Run(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("run = %+v", res)
	}
	if res.CursorAdvanced {
		t.Fatal("inconclusive must not advance cursor")
	}
	m, _ := LoadManifest(SessionDir(state, "marcus-skip-inconclusive"))
	if m != nil && strings.TrimSpace(m.LastSinkMaxMessageTimestamp) != "" {
		t.Fatalf("cursor should stay empty, got %q", m.LastSinkMaxMessageTimestamp)
	}
	if m != nil && strings.TrimSpace(m.LastSinkAt) != "" {
		t.Fatalf("inconclusive must not set last_sink_at (keep sinkable), got %q", m.LastSinkAt)
	}
	st := BuildStatus(m, t0, 1, true, "")
	if st == nil || !IsAutoSinkable(st) || st.State != StateReady {
		t.Fatalf("inconclusive should stay auto-sinkable ready, got %+v", st)
	}
}

func TestRun_CreateMRZeroTipFallsBackCursorToLastSinkAt(t *testing.T) {
	state := t.TempDir()
	hub, _ := setupHubRemote(t)
	runnerDir := t.TempDir()
	now := time.Date(2026, 8, 25, 14, 0, 0, 0, time.UTC)
	opts := Opts{
		StateDir:            state,
		HubDir:              hub,
		SessionID:           "marcus-zero-tip",
		Mode:                ModeHeadless,
		CreateMR:            true,
		AutoMergeMR:         true,
		ResolveFn:           func(string) (string, string, error) { return "grok-tty", "g-zero", nil },
		ResolveSessionDirFn: func(string, string) (string, error) { return runnerDir, nil },
		// Messages exist for delta, but timestamps missing → tip stays zero.
		MessagesFn: func(string, string, *sessions.MessagesOpts) (*sessions.MessagesResult, error) {
			return &sessions.MessagesResult{
				Total: 1,
				Messages: []sessions.ChatMessage{
					{Kind: sessions.MessageKindUser, Text: "hi"},
				},
			}, nil
		},
		AgentFn: func(_ context.Context, o agentrunapi.RunOpts, _ string) (string, error) {
			body, _ := json.Marshal(ShipResult{
				HasNewKnowledges: BoolPtr(false),
				SkipReason:       SkipReasonNoNew,
			})
			return "", os.WriteFile(o.ResultFile, body, 0o644)
		},
		NowFn: func() time.Time { return now },
	}
	res, err := Run(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || !res.CursorAdvanced {
		t.Fatalf("want ok+cursor advanced on zero tip, got %+v", res)
	}
	m, err := LoadManifest(SessionDir(state, "marcus-zero-tip"))
	if err != nil || m == nil {
		t.Fatalf("manifest: %+v err=%v", m, err)
	}
	if strings.TrimSpace(m.LastSinkAt) == "" || m.LastSinkMaxMessageTimestamp != m.LastSinkAt {
		t.Fatalf("cursor should fall back to last_sink_at: at=%q cursor=%q", m.LastSinkAt, m.LastSinkMaxMessageTimestamp)
	}
	st := BuildStatus(m, time.Time{}, 1, true, "")
	if st == nil || IsAutoSinkable(st) || st.State != StateSunk {
		t.Fatalf("after zero-tip finish want sunk, got %+v", st)
	}
}
