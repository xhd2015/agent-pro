package knowledgesink

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func TestShipToMR_AllowsDeletedTrackedFile(t *testing.T) {
	hub, bare := setupHubRemote(t)
	if err := os.Remove(filepath.Join(hub, "README.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "topics", "kept.md"), []byte("k\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ship := &ShipResult{
		GitCommitMsg:  "docs(kb): drop README keep topic",
		GitBranchName: "tester/2026-03-24-drop-readme",
		GitCommitFiles: ShipCommitFiles{
			Add:    []string{"topics/kept.md"},
			Delete: []string{"README.md"},
		},
	}
	res, err := ShipToMR(Opts{}, hub, ship, true)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Merged {
		t.Fatalf("expected merged: %+v", res)
	}
	// Deletion landed on master.
	if out, err := exec.Command("git", "-C", bare, "cat-file", "-e", "master:README.md").CombinedOutput(); err == nil {
		t.Fatalf("README.md should be deleted on master: %s", out)
	}
	if out, err := exec.Command("git", "-C", bare, "cat-file", "-e", "master:topics/kept.md").CombinedOutput(); err != nil {
		t.Fatalf("kept.md missing on master: %s %v", out, err)
	}
}

func TestShipToMR_VerboseNotices(t *testing.T) {
	hub, _ := setupHubRemote(t)
	if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "topics", "v.md"), []byte("v\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	ship := &ShipResult{
		GitCommitMsg:  "docs(kb): verbose ship",
		GitBranchName: "tester/2026-03-24-verbose-ship",
		GitCommitFiles: ShipCommitFiles{
			Add: []string{"topics/v.md"},
		},
	}
	_, err := ShipToMR(Opts{Verbose: true, Stderr: &stderr}, hub, ship, true)
	if err != nil {
		t.Fatal(err)
	}
	got := stderr.String()
	for _, want := range []string{
		"notice: ship: stash",
		"notice: ship: checkout -B",
		"notice: ship: git commit",
		"notice: ship: git push",
		"notice: ship: auto-merge",
		"notice: ship: done",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("stderr missing %q:\n%s", want, got)
		}
	}
}

func TestShipToMR_CreateAndAutoMerge(t *testing.T) {
	hub, bare := setupHubRemote(t)

	opts := Opts{}
	ship := &ShipResult{
		GitCommitMsg:  "docs(kb): add topic",
		GitBranchName: "tester/2026-03-24-add-topic",
		GitCommitFiles: ShipCommitFiles{
			Add: []string{"topics/new.md"},
		},
	}
	if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "topics", "new.md"), []byte("# new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := ShipToMR(opts, hub, ship, true)
	if err != nil {
		t.Fatal(err)
	}
	if res.Branch != ship.GitBranchName || res.Commit == "" {
		t.Fatalf("ship = %+v", res)
	}
	if !res.Merged {
		t.Fatalf("expected merged: %+v", res)
	}
	// Branch on remote
	out, err := exec.Command("git", "-C", bare, "rev-parse", ship.GitBranchName).CombinedOutput()
	if err != nil {
		t.Fatalf("bare branch: %s %v", out, err)
	}
	master, err := exec.Command("git", "-C", bare, "rev-parse", "master").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(out)) != strings.TrimSpace(string(master)) {
		t.Fatalf("master tip %s != branch tip %s", master, out)
	}
	// Fallback URL built when push options unsupported on bare remote
	if res.MRURL == "" && res.Warning == "" {
		// bare repos ignore -o; we should have fallen back
		t.Logf("mr_url=%q warning=%q push_option=%v", res.MRURL, res.Warning, res.PushOption)
	}
}

func TestRun_CreateMR_Happy(t *testing.T) {
	state := t.TempDir()
	hub, _ := setupHubRemote(t)
	runnerDir := t.TempDir()
	t0 := time.Date(2026, 3, 20, 10, 0, 0, 0, time.Local)
	t1 := t0.Add(time.Hour)

	opts := Opts{
		StateDir:            state,
		HubDir:              hub,
		SessionID:           "marcus-mr",
		Mode:                ModeHeadless,
		CreateMR:            true,
		AutoMergeMR:         true,
		NowFn:               func() time.Time { return t1 },
		ResolveFn:           func(string) (string, string, error) { return "grok-tty", "grok-1", nil },
		ResolveSessionDirFn: func(string, string) (string, error) { return runnerDir, nil },
		MessagesFn: func(string, string, *sessions.MessagesOpts) (*sessions.MessagesResult, error) {
			return &sessions.MessagesResult{
				Total: 2,
				Messages: []sessions.ChatMessage{
					{Kind: sessions.MessageKindUser, Text: "a", Timestamp: t0},
					{Kind: sessions.MessageKindResponse, Text: "b", Timestamp: t1},
				},
			}, nil
		},
		AgentFn: func(_ context.Context, o agentrunapi.RunOpts, schema string) (string, error) {
			if schema != "" {
				t.Fatal("create-mr must not use propose RunJSON schema")
			}
			if !strings.Contains(o.Prompt, "## Output") || !strings.Contains(o.Prompt, "result.json") {
				t.Fatalf("prompt missing Output: %s", o.Prompt)
			}
			resultPath := resultJSONAbsPath(SessionDir(state, "marcus-mr"), 0)
			if o.ResultFile != resultPath {
				t.Fatalf("ResultFile = %q want %q", o.ResultFile, resultPath)
			}
			if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
				return "", err
			}
			if err := os.WriteFile(filepath.Join(hub, "topics", "from-agent.md"), []byte("# from agent\n"), 0o644); err != nil {
				return "", err
			}
			body, _ := json.Marshal(ShipResult{
				GitCommitMsg:  "docs(kb): from agent",
				GitBranchName: "tester/2026-03-20-from-agent",
				GitCommitFiles: ShipCommitFiles{
					Add: []string{"topics/from-agent.md"},
				},
			})
			if err := os.WriteFile(resultPath, body, 0o644); err != nil {
				return "", err
			}
			return "", nil
		},
	}

	res, err := Run(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || !res.CreateMR || !res.Merged {
		t.Fatalf("run = %+v", res)
	}
	if !res.CursorAdvanced {
		t.Fatalf("cursor should advance: %+v", res)
	}
	man, _ := LoadManifest(SessionDir(state, "marcus-mr"))
	if man == nil || man.LastSinkMaxMessageTimestamp == "" {
		t.Fatalf("manifest cursor: %+v", man)
	}
	if man.LastBranch == "" || man.LastCommit == "" {
		t.Fatalf("manifest ship fields: %+v", man)
	}
}

func TestRun_CreateMR_AllowsDirtyHub(t *testing.T) {
	state := t.TempDir()
	hub, _ := setupHubRemote(t)
	if err := os.WriteFile(filepath.Join(hub, "unrelated-dirt.md"), []byte("leave me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runnerDir := t.TempDir()
	t1 := time.Date(2026, 3, 20, 11, 0, 0, 0, time.Local)

	opts := Opts{
		StateDir:            state,
		HubDir:              hub,
		SessionID:           "marcus-dirty-ok",
		Mode:                ModeHeadless,
		CreateMR:            true,
		AutoMergeMR:         true,
		NowFn:               func() time.Time { return t1 },
		ResolveFn:           func(string) (string, string, error) { return "grok-tty", "grok-1", nil },
		ResolveSessionDirFn: func(string, string) (string, error) { return runnerDir, nil },
		MessagesFn: func(string, string, *sessions.MessagesOpts) (*sessions.MessagesResult, error) {
			return &sessions.MessagesResult{
				Total: 1,
				Messages: []sessions.ChatMessage{
					{Kind: sessions.MessageKindUser, Text: "x", Timestamp: t1},
				},
			}, nil
		},
		AgentFn: func(_ context.Context, o agentrunapi.RunOpts, schema string) (string, error) {
			if schema != "" {
				t.Fatal("create-mr must not use propose RunJSON schema")
			}
			want := resultJSONAbsPath(SessionDir(state, "marcus-dirty-ok"), 0)
			if o.ResultFile != want {
				t.Fatalf("ResultFile = %q want %q", o.ResultFile, want)
			}
			if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
				return "", err
			}
			if err := os.WriteFile(filepath.Join(hub, "topics", "shipped.md"), []byte("# shipped\n"), 0o644); err != nil {
				return "", err
			}
			body, _ := json.Marshal(ShipResult{
				GitCommitMsg:  "docs(kb): shipped despite dirt",
				GitBranchName: "tester/2026-03-20-shipped-dirt",
				GitCommitFiles: ShipCommitFiles{
					Add: []string{"topics/shipped.md"},
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
	if !res.OK {
		t.Fatalf("dirty hub should be allowed: %+v", res)
	}
	if _, err := os.Stat(filepath.Join(hub, "unrelated-dirt.md")); err != nil {
		t.Fatalf("unrelated dirt should remain: %v", err)
	}
}

func setupHubRemote(t *testing.T) (hub, bare string) {
	t.Helper()
	bare = t.TempDir()
	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %s (%v)", args, out, err)
		}
	}
	run(bare, "init", "--bare", "-b", "master")

	hub = t.TempDir()
	run(hub, "init", "-b", "master")
	run(hub, "config", "user.email", "tester@example.com")
	run(hub, "config", "user.name", "Tester")
	if err := os.WriteFile(filepath.Join(hub, "SINK.md"), []byte("# sink\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "README.md"), []byte("hub\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(hub, "add", ".")
	run(hub, "commit", "-m", "init")
	run(hub, "remote", "add", "origin", bare)
	run(hub, "push", "-u", "origin", "master")
	return hub, bare
}
