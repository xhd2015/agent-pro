package knowledgesink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/sweep"
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
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): drop README keep topic",
		GitBranchName:    "tester/2026-03-24-drop-readme",
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
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): verbose ship",
		GitBranchName:    "tester/2026-03-24-verbose-ship",
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
		"notice: ship: commit on",
		"notice: ship: git commit",
		"notice: ship: git push origin ",
		"notice: ship: ff origin/master",
		"notice: ship: done",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("stderr missing %q:\n%s", want, got)
		}
	}
	for _, ban := range []string{
		"checkout -B",
		"ship: stash",
	} {
		if strings.Contains(got, ban) {
			t.Fatalf("stderr should not contain %q:\n%s", ban, got)
		}
	}
}

func TestShipToMR_CreateAndAutoMerge(t *testing.T) {
	hub, bare := setupHubRemote(t)

	opts := Opts{}
	ship := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): add topic",
		GitBranchName:    "tester/2026-03-24-add-topic",
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
				HasNewKnowledges: BoolPtr(true),
				GitCommitMsg:     "docs(kb): from agent",
				GitBranchName:    "tester/2026-03-20-from-agent",
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
	if man == nil || SunkCursor(man) == "" || CheckedCursor(man) == "" {
		t.Fatalf("ship must advance sunk+checked: %+v", man)
	}
	if SunkCursor(man) != CheckedCursor(man) {
		t.Fatalf("ship cursors should match: sunk=%q checked=%q", SunkCursor(man), CheckedCursor(man))
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
				HasNewKnowledges: BoolPtr(true),
				GitCommitMsg:     "docs(kb): shipped despite dirt",
				GitBranchName:    "tester/2026-03-20-shipped-dirt",
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

func gitHub(t *testing.T, hub string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", hub}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %s (%v)", args, out, err)
	}
	return strings.TrimSpace(string(out))
}

func TestShipToMR_KeepsCurrentBranch(t *testing.T) {
	hub, _ := setupHubRemote(t)
	gitHub(t, hub, "checkout", "-B", "marcus-workspace")
	gitHub(t, hub, "branch", "--set-upstream-to=origin/master", "marcus-workspace")

	if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "topics", "stay.md"), []byte("stay\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Unrelated dirt must survive.
	if err := os.WriteFile(filepath.Join(hub, "unrelated.md"), []byte("leave\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ship := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): stay on workspace branch",
		GitBranchName:    "tester/2026-03-24-stay-branch",
		GitCommitFiles: ShipCommitFiles{
			Add: []string{"topics/stay.md"},
		},
	}
	res, err := ShipToMR(Opts{}, hub, ship, true)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Merged {
		t.Fatalf("expected merged: %+v", res)
	}
	if got := gitHub(t, hub, "branch", "--show-current"); got != "marcus-workspace" {
		t.Fatalf("current branch = %q, want marcus-workspace", got)
	}
	if _, err := os.Stat(filepath.Join(hub, "unrelated.md")); err != nil {
		t.Fatalf("unrelated dirt missing: %v", err)
	}
}

// A hub whose HEAD is ahead of origin/master used to be refused outright, which
// made every diverged workspace permanently unable to sink. It is now merged: the
// pre-existing local commit travels with the ship, and the log says so.
func TestShipToMR_MergesDivergedHEAD(t *testing.T) {
	hub, bare := setupHubRemote(t)
	if err := os.WriteFile(filepath.Join(hub, "local-only.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitHub(t, hub, "add", "local-only.md")
	gitHub(t, hub, "commit", "-m", "local ahead of origin")
	diverged := gitHub(t, hub, "rev-parse", "HEAD")

	if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "topics", "ship.md"), []byte("s\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ship := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): ship from a diverged hub",
		GitBranchName:    "tester/2026-03-24-diverged",
		GitCommitFiles: ShipCommitFiles{
			Add: []string{"topics/ship.md"},
		},
	}
	res, err := ShipToMR(Opts{}, hub, ship, true)
	if err != nil {
		t.Fatalf("a diverged hub must merge, not fail: %v", err)
	}
	if !res.Merged {
		t.Fatalf("auto-merge did not run: %+v", res)
	}
	// The pre-existing commit is published as part of the same tip, and the count
	// is reported rather than left for the reviewer to discover in the MR diff.
	if res.Foreign != 1 {
		t.Fatalf("foreign = %d, want 1", res.Foreign)
	}
	masterTip := gitHub(t, bare, "rev-parse", "master")
	if masterTip != res.Tip {
		t.Fatalf("master = %s, want the published tip %s", masterTip, res.Tip)
	}
	if masterTip == diverged {
		t.Fatal("master must be the new tip, not the pre-ship commit")
	}
	// The feature branch points at the same commit as master: anything else means
	// the MR would review a commit that was never published to the target.
	featureTip := gitHub(t, bare, "rev-parse", "tester/2026-03-24-diverged")
	if featureTip != masterTip {
		t.Fatalf("feature branch = %s, master = %s; they must agree", featureTip, masterTip)
	}
	if got := gitHub(t, hub, "branch", "--show-current"); got != "master" {
		t.Fatalf("branch changed to %q", got)
	}
}

func TestShipToMR_IgnoresFailingHooks(t *testing.T) {
	hub, _ := setupHubRemote(t)
	hooks := t.TempDir()
	for _, name := range []string{"post-checkout", "post-commit", "pre-push"} {
		path := filepath.Join(hooks, name)
		body := "#!/bin/sh\necho hook-fail-" + name + " >&2\nexit 2\n"
		if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	gitHub(t, hub, "config", "core.hooksPath", hooks)

	if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "topics", "hooks.md"), []byte("h\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ship := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): ignore hooks",
		GitBranchName:    "tester/2026-03-24-ignore-hooks",
		GitCommitFiles: ShipCommitFiles{
			Add: []string{"topics/hooks.md"},
		},
	}
	res, err := ShipToMR(Opts{}, hub, ship, true)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Merged {
		t.Fatalf("expected merged: %+v", res)
	}
}

func TestShipToMR_AutoMergeWithMasterLockedByWorktree(t *testing.T) {
	hub, bare := setupHubRemote(t)
	gitHub(t, hub, "checkout", "-B", "marcus-workspace")
	gitHub(t, hub, "branch", "--set-upstream-to=origin/master", "marcus-workspace")

	// Primary-style worktree holds master (cannot checkout master in hub).
	other := t.TempDir()
	cmd := exec.Command("git", "-C", hub, "worktree", "add", other, "master")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("worktree add: %s (%v)", out, err)
	}
	t.Cleanup(func() {
		_ = exec.Command("git", "-C", hub, "worktree", "remove", "--force", other).Run()
	})

	if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "topics", "wt.md"), []byte("wt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ship := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): worktree-safe merge",
		GitBranchName:    "tester/2026-03-24-wt-merge",
		GitCommitFiles: ShipCommitFiles{
			Add: []string{"topics/wt.md"},
		},
	}
	res, err := ShipToMR(Opts{}, hub, ship, true)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Merged {
		t.Fatalf("expected merged: %+v", res)
	}
	if got := gitHub(t, hub, "branch", "--show-current"); got != "marcus-workspace" {
		t.Fatalf("current branch = %q", got)
	}
	master, err := exec.Command("git", "-C", bare, "rev-parse", "master").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	feature, err := exec.Command("git", "-C", bare, "rev-parse", ship.GitBranchName).CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(master)) != strings.TrimSpace(string(feature)) {
		t.Fatalf("master %s != feature %s", master, feature)
	}
}

func TestGitErrMsgDropsSwitchedNoise(t *testing.T) {
	stderr := "Switched to a new branch 'x'\n\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path.\n"
	got := gitErrMsg(stderr, fmt.Errorf("exit status 1"))
	if !strings.Contains(got, "git-lfs") {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(got, "Switched to a new branch") {
		t.Fatalf("should drop switch noise: %q", got)
	}
}

// The fail mode this whole change exists for: the agent writes knowledge, the hub
// has unrelated dirt, and that dirt stays behind. The next `marcus kb` then skips
// with "dirty working tree" and nobody sees a new KB until someone notices.
func TestShipToMR_CarriesOtherCommittableDirt(t *testing.T) {
	hub, bare := setupHubRemote(t)
	if err := os.WriteFile(filepath.Join(hub, "unrelated.md"), []byte("leftover\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "topics", "shipped.md"), []byte("# shipped\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	png := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d}
	if err := os.WriteFile(filepath.Join(hub, "logo.png"), png, 0o644); err != nil {
		t.Fatal(err)
	}
	ship := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): ship and sweep",
		GitBranchName:    "tester/2026-03-25-sweep",
		GitCommitFiles:   ShipCommitFiles{Add: []string{"topics/shipped.md"}},
	}
	res, err := ShipToMR(Opts{
		SweepPlanFn: func(dir string) (*SweepPlan, error) {
			return sweep.BuildPlan(context.Background(), sweep.Options{Dir: dir})
		},
	}, hub, ship, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Swept) != 1 || res.Swept[0] != "unrelated.md" {
		t.Fatalf("swept = %v, want the unrelated text file", res.Swept)
	}
	// The binary is left alone but must be reported, not silently dropped.
	if len(res.Skipped) != 1 || res.Skipped[0].Path != "logo.png" {
		t.Fatalf("skipped = %+v", res.Skipped)
	}
	if strings.TrimSpace(gitHub(t, bare, "ls-tree", "-r", "--name-only", "master")) == "" {
		t.Fatal("nothing reached master")
	}
	if tracked := gitHub(t, bare, "ls-tree", "-r", "--name-only", "master"); strings.Contains(tracked, "logo.png") {
		t.Fatalf("binary must not be published:\n%s", tracked)
	}
	if _, err := os.Stat(filepath.Join(hub, "logo.png")); err != nil {
		t.Fatalf("binary must stay on disk: %v", err)
	}
}

// The ordering rule: when the hub's dirt touches the same file the incoming commit
// touches, git refuses to merge before it detects a conflict. Committing first is
// what turns that dead end into a resolvable conflict.
func TestShipToMR_CommitsBeforeMergingSoOverlapIsResolvable(t *testing.T) {
	hub, _ := setupHubRemote(t)
	if err := os.WriteFile(filepath.Join(hub, "INDEX.md"), []byte("# index\nbase\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitHub(t, hub, "add", "INDEX.md")
	gitHub(t, hub, "commit", "-m", "index")
	gitHub(t, hub, "push", "origin", "HEAD:master")

	// A peer changes the same file upstream.
	peer := filepath.Join(t.TempDir(), "peer")
	gitHub(t, filepath.Dir(peer), "clone", "-q", gitHub(t, hub, "remote", "get-url", "origin"), peer)
	gitHub(t, peer, "config", "user.email", "p@example.com")
	gitHub(t, peer, "config", "user.name", "p")
	gitHub(t, peer, "config", "core.hooksPath", os.DevNull)
	if err := os.WriteFile(filepath.Join(peer, "INDEX.md"), []byte("# index\npeer\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitHub(t, peer, "add", "INDEX.md")
	gitHub(t, peer, "commit", "-m", "peer edit")
	gitHub(t, peer, "push", "-q", "origin", "HEAD:master")

	// The hub edits the same file without committing it, plus ships new knowledge.
	if err := os.WriteFile(filepath.Join(hub, "INDEX.md"), []byte("# index\nmine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "topics", "t.md"), []byte("# t\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ship := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): overlapping edit",
		GitBranchName:    "tester/2026-03-25-overlap",
		GitCommitFiles:   ShipCommitFiles{Add: []string{"topics/t.md"}},
	}
	// The resolver runs only if the commit happened first; if it does not run, the
	// merge either fails (dirty overlap) or leaves a conflict behind.
	resolved := 0
	agent := AgentFunc(func(_ context.Context, o agentrunapi.RunOpts, _ string) (string, error) {
		resolved++
		// Both sides survive, which is the KB rule for an index page.
		return "", os.WriteFile(filepath.Join(o.WorkspaceDir, "INDEX.md"), []byte("# index\nmine\npeer\n"), 0o644)
	})
	res, err := ShipToMR(Opts{
		AgentFn:       agent,
		ResolvePrompt: "resolve the conflicts",
		SweepPlanFn: func(dir string) (*SweepPlan, error) {
			return sweep.BuildPlan(context.Background(), sweep.Options{Dir: dir})
		},
	}, hub, ship, true)
	if err != nil {
		t.Fatalf("overlapping dirt must be committed then merged, not refused: %v", err)
	}
	if resolved != 1 {
		t.Fatalf("resolver ran %d time(s), want 1", resolved)
	}
	if len(res.Resolved) != 1 || res.Resolved[0] != "INDEX.md" {
		t.Fatalf("resolved = %v", res.Resolved)
	}
	body, _ := os.ReadFile(filepath.Join(hub, "INDEX.md"))
	if !strings.Contains(string(body), "mine") || !strings.Contains(string(body), "peer") {
		t.Fatalf("both sides must survive:\n%s", body)
	}
	if strings.Contains(string(body), "<<<<<<<") {
		t.Fatalf("markers left behind:\n%s", body)
	}
	// The swept path travelled with the commit rather than blocking the merge.
	if len(res.Swept) != 1 || res.Swept[0] != "INDEX.md" {
		t.Fatalf("swept = %v, want INDEX.md", res.Swept)
	}
	if !res.Merged {
		t.Fatal("ship did not publish")
	}
}

// The whole point of the ordered ship: a stale hub plus a peer commit plus a
// conflict on the same page must still end with master published and the MR branch
// pointing at the same commit.
func TestShipToMR_DivergedWithConflictPublishesTheSameTipEverywhere(t *testing.T) {
	hub, bare := setupHubRemote(t)
	if err := os.WriteFile(filepath.Join(hub, "INDEX.md"), []byte("# index\nbase\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitHub(t, hub, "add", "INDEX.md")
	gitHub(t, hub, "commit", "-m", "index")
	gitHub(t, hub, "push", "origin", "HEAD:master")

	// A peer publishes a change to the same page.
	peer := filepath.Join(t.TempDir(), "peer")
	gitHub(t, filepath.Dir(peer), "clone", "-q", gitHub(t, hub, "remote", "get-url", "origin"), peer)
	for _, args := range [][]string{
		{"config", "user.email", "p@example.com"},
		{"config", "user.name", "p"},
		{"config", "core.hooksPath", os.DevNull},
	} {
		gitHub(t, peer, args...)
	}
	if err := os.WriteFile(filepath.Join(peer, "INDEX.md"), []byte("# index\npeer\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitHub(t, peer, "add", "INDEX.md")
	gitHub(t, peer, "commit", "-m", "peer edit")
	gitHub(t, peer, "push", "-q", "origin", "HEAD:master")

	// The hub has its own unpushed commit plus an uncommitted edit to the same file.
	if err := os.WriteFile(filepath.Join(hub, "local.md"), []byte("# local\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitHub(t, hub, "add", "local.md")
	gitHub(t, hub, "commit", "-m", "local commit")
	if err := os.WriteFile(filepath.Join(hub, "INDEX.md"), []byte("# index\nmine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "topics", "t.md"), []byte("# t\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ship := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): diverged with conflict",
		GitBranchName:    "tester/2026-03-26-diverged-conflict",
		GitCommitFiles:   ShipCommitFiles{Add: []string{"topics/t.md"}},
	}
	agent := AgentFunc(func(_ context.Context, o agentrunapi.RunOpts, _ string) (string, error) {
		return "", os.WriteFile(filepath.Join(o.WorkspaceDir, "INDEX.md"), []byte("# index\nmine\npeer\n"), 0o644)
	})
	res, err := ShipToMR(Opts{
		AgentFn: agent, ResolvePrompt: "resolve",
		SweepPlanFn: func(dir string) (*SweepPlan, error) {
			return sweep.BuildPlan(context.Background(), sweep.Options{Dir: dir})
		},
	}, hub, ship, true)
	if err != nil {
		t.Fatalf("%v (result %+v)", err, res)
	}
	master := gitHub(t, bare, "rev-parse", "master")
	feature := gitHub(t, bare, "rev-parse", "tester/2026-03-26-diverged-conflict")
	// The invariant: one tip, both refs. Pushing the pre-merge commit would make
	// master lag behind the MR branch.
	if master != feature {
		t.Fatalf("master = %s, feature = %s; both must be the published tip", master, feature)
	}
	if res.Tip != master {
		t.Fatalf("reported tip %s != master %s", res.Tip, master)
	}
	if len(res.Resolved) != 1 || res.Resolved[0] != "INDEX.md" {
		t.Fatalf("resolved = %v", res.Resolved)
	}
	if len(res.Swept) != 1 || res.Swept[0] != "INDEX.md" {
		t.Fatalf("swept = %v", res.Swept)
	}
	body, _ := os.ReadFile(filepath.Join(hub, "INDEX.md"))
	if !strings.Contains(string(body), "mine") || !strings.Contains(string(body), "peer") {
		t.Fatalf("both sides must survive:\n%s", body)
	}
}

// A peer publishing between our merge and our ff-push makes the target reject us.
// The retry must re-merge and publish the new tip to both refs -- and must not push
// the pre-merge commit, which the target would reject forever.
func TestShipToMR_RetriesWhenTargetMovesDuringShip(t *testing.T) {
	hub, bare := setupHubRemote(t)
	hooksDir := filepath.Join(t.TempDir(), "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(hooksDir, "fired")
	peerScript := filepath.Join(t.TempDir(), "peer-push.sh")
	script := "#!/bin/sh\nset -e\n" +
		"d=$(mktemp -d)\n" +
		"git clone --quiet " + bare + " \"$d\"\n" +
		"cd \"$d\"\n" +
		"git config user.email peer@example.com\n" +
		"git config user.name peer\n" +
		"git config core.hooksPath /dev/null\n" +
		"echo peer > raced.md\n" +
		"git add -A\n" +
		"git commit -qm raced\n" +
		"git push -q origin HEAD:master\n"
	if err := os.WriteFile(peerScript, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	// The hook fires on the first push only: afterwards the retry must succeed.
	pushHook := "#!/bin/sh\nif [ ! -f " + marker + " ]; then\n  touch " + marker + "\n  " + peerScript + "\nfi\n"
	if err := os.WriteFile(filepath.Join(hooksDir, "pre-push"), []byte(pushHook), 0o755); err != nil {
		t.Fatal(err)
	}
	gitHub(t, hub, "config", "core.hooksPath", hooksDir)
	t.Cleanup(func() { _ = os.WriteFile(marker, []byte("x"), 0o644) })

	if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "topics", "r.md"), []byte("# r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ship := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): retry publish",
		GitBranchName:    "tester/2026-03-27-retry",
		GitCommitFiles:   ShipCommitFiles{Add: []string{"topics/r.md"}},
	}
	// withShipGit disables hooks for its own git calls, so the race must be induced
	// on the ship's push path instead: allow hooks for this call.
	res, err := ShipToMR(Opts{HonorHooks: true, PublishAttempts: 3}, hub, ship, true)
	if err != nil {
		t.Fatalf("%v (result %+v)", err, res)
	}
	master := gitHub(t, bare, "rev-parse", "master")
	feature := gitHub(t, bare, "rev-parse", "tester/2026-03-27-retry")
	if master != feature {
		t.Fatalf("master = %s, feature = %s; the retry must publish one tip to both", master, feature)
	}
	if res.Tip != master {
		t.Fatalf("reported tip %s != master %s", res.Tip, master)
	}
	if res.Attempts < 2 {
		t.Fatalf("attempts = %d; the race must have been retried", res.Attempts)
	}
	// The racing commit must be merged in, not raced past.
	if _, statErr := os.Stat(filepath.Join(hub, "raced.md")); statErr != nil {
		t.Fatalf("the racing commit must be merged: %v", statErr)
	}
}
