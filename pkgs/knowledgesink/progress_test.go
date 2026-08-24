package knowledgesink

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func TestProgressFormat(t *testing.T) {
	var buf bytes.Buffer
	progress(Opts{Stderr: &buf}, 2, 4, stageAgent, "wait result.json")
	got := buf.String()
	want := "[2/4] agent        wait result.json\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRun_ProgressStagesCreateMR(t *testing.T) {
	state := t.TempDir()
	hub, _ := setupHubRemote(t)
	runnerDir := t.TempDir()
	t0 := time.Now()
	var stderr bytes.Buffer
	opts := Opts{
		StateDir:            state,
		HubDir:              hub,
		SessionID:           "marcus-progress",
		Mode:                ModeHeadless,
		CreateMR:            true,
		AutoMergeMR:         true,
		Stderr:              &stderr,
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
		AgentFn: func(_ context.Context, _ agentrunapi.RunOpts, _ string) (string, error) {
			if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
				return "", err
			}
			if err := os.WriteFile(filepath.Join(hub, "topics", "p.md"), []byte("p\n"), 0o644); err != nil {
				return "", err
			}
			want := resultJSONAbsPath(SessionDir(state, "marcus-progress"), 0)
			body, _ := json.Marshal(ShipResult{
				GitCommitMsg:  "docs(kb): progress",
				GitBranchName: "tester/2026-03-24-progress",
				GitCommitFiles: ShipCommitFiles{
					Add: []string{"topics/p.md"},
				},
			})
			return "", os.WriteFile(want, body, 0o644)
		},
	}
	res, err := Run(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("run = %+v", res)
	}
	got := stderr.String()
	for _, want := range []string{
		"[1/5] launch",
		"[2/5] agent",
		"[3/5] validate",
		"[4/5] ship",
		"[5/5] merge",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("stderr missing %q:\n%s", want, got)
		}
	}
	// Without Verbose, no fine-grained ship: git lines.
	if strings.Contains(got, "notice: ship: git commit") {
		t.Fatalf("non-verbose should not dump git micro-steps:\n%s", got)
	}
}

func TestRun_DryRunProgressStages(t *testing.T) {
	state := t.TempDir()
	hub := filepath.Join(t.TempDir(), "knowledge-base-hub")
	if err := os.MkdirAll(hub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "SINK.md"), []byte("#\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	opts := Opts{
		StateDir:            state,
		HubDir:              hub,
		SessionID:           "marcus-dry-prog",
		DryRun:              true,
		CreateMR:            true,
		Stderr:              &stderr,
		ResolveFn:           func(string) (string, string, error) { return "grok-tty", "g1", nil },
		ResolveSessionDirFn: func(string, string) (string, error) { return t.TempDir(), nil },
		MessagesFn: func(string, string, *sessions.MessagesOpts) (*sessions.MessagesResult, error) {
			return &sessions.MessagesResult{Total: 0}, nil
		},
	}
	res, err := Run(context.Background(), opts)
	if err != nil || !res.OK || !res.DryRun {
		t.Fatalf("%v %+v", err, res)
	}
	got := stderr.String()
	if !strings.Contains(got, "[1/4] launch") || !strings.Contains(got, "skip (dry-run)") {
		t.Fatalf("dry-run stages:\n%s", got)
	}
}
