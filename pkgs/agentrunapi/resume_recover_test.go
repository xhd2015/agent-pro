package agentrunapi

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func TestLocalGrokRunnerSessionMissing_noUpdatesJSONL(t *testing.T) {
	grokHome := t.TempDir()
	t.Setenv("GROK_HOME", grokHome)
	t.Setenv("HOME", t.TempDir()) // avoid leaking real ~/.grok

	opts := Opts{AgentRunner: "grok-tty", WorkspaceDir: "/tmp/ws"}
	meta := agentstorage.SessionMeta{
		Runner:          "grok-tty",
		RunnerSessionID: "01a01e5d-a7eb-7d31-b013-04f57284959f",
		Workspace:       "/tmp/ws",
	}
	if !localGrokRunnerSessionMissing(opts, meta) {
		t.Fatal("expected missing when no updates.jsonl under GROK_HOME")
	}

	// Create local session data → not missing.
	dir := filepath.Join(grokHome, "sessions", "%2Ftmp%2Fws", meta.RunnerSessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), []byte("{}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if localGrokRunnerSessionMissing(opts, meta) {
		t.Fatal("expected present after updates.jsonl exists")
	}
}

func TestLocalGrokRunnerSessionMissing_nonGrokSkipped(t *testing.T) {
	opts := Opts{AgentRunner: "codex-tty"}
	meta := agentstorage.SessionMeta{
		Runner:          "codex-tty",
		RunnerSessionID: "019fdca1-3893-7fa3-a8aa-ebc1ccc750a0",
	}
	if localGrokRunnerSessionMissing(opts, meta) {
		t.Fatal("codex must not use grok local-session precheck")
	}
}

func TestAutoSendOrResume_missingLocalGrokClearsBindAndModeRuns(t *testing.T) {
	home := t.TempDir()
	grokHome := t.TempDir()
	t.Setenv("GROK_HOME", grokHome)
	t.Setenv("HOME", t.TempDir())

	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	const sessionID = "seatalk-orphan-1"
	const staleID = "01a01e5d-dead-beef-b013-04f57284959f"
	meta := agentstorage.SessionMeta{
		SessionID:       sessionID,
		Runner:          "grok-tty",
		RunnerSessionID: staleID,
		Status:          "exited",
		Workspace:       "/tmp/ws",
	}
	if err := store.CreateSession(sessionID, meta); err != nil {
		t.Fatal(err)
	}
	// Durable bind.json that should be removed on recover.
	bindDir := filepath.Join(home, "sessions", sessionID)
	if err := os.WriteFile(filepath.Join(bindDir, "bind.json"), []byte(`{"state":"ok","runner_session_id":"`+staleID+`"}`), 0644); err != nil {
		t.Fatal(err)
	}

	var resumeN, runN int
	var stderr bytes.Buffer
	exited := true
	err = AutoSendOrResume(context.Background(), Opts{
		SessionID:   sessionID,
		Prompt:      "follow up",
		AgentRunner: "grok-tty",
		Store:       store,
		Stderr:      &stderr,
		Probe: func(store agentstorage.Store, m agentstorage.SessionMeta) (ProbeReport, error) {
			// Pretend Classify would ModeResume (bound + exited).
			if strings.TrimSpace(m.RunnerSessionID) == "" {
				return ProbeReport{}, nil
			}
			return ProbeReport{ResumeReady: true, RunnerExited: &exited}, nil
		},
		ResumeSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta) error {
			resumeN++
			return fmt.Errorf("resume should not be called when local grok data is missing")
		},
		RunSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta, found bool) error {
			runN++
			if strings.TrimSpace(m.RunnerSessionID) != "" {
				t.Fatalf("ModeRun meta should be unbound, got %q", m.RunnerSessionID)
			}
			if !found {
				t.Fatal("expected found=true for existing agent-run session")
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("AutoSendOrResume: %v", err)
	}
	if resumeN != 0 {
		t.Fatalf("ResumeSession called %d times; want 0", resumeN)
	}
	if runN != 1 {
		t.Fatalf("RunSession called %d times; want 1", runN)
	}
	if !strings.Contains(stderr.String(), staleID) || !strings.Contains(stderr.String(), "no local grok data") {
		t.Fatalf("stderr should warn about missing local data, got %q", stderr.String())
	}
	sess, gerr := store.GetSession(sessionID)
	if gerr != nil {
		t.Fatal(gerr)
	}
	if strings.TrimSpace(sess.Meta.RunnerSessionID) != "" {
		t.Fatalf("store still bound to %q", sess.Meta.RunnerSessionID)
	}
	if _, err := os.Stat(filepath.Join(bindDir, "bind.json")); !os.IsNotExist(err) {
		t.Fatalf("bind.json should be removed, err=%v", err)
	}
}

func TestAutoSendOrResume_localGrokPresentStillResumes(t *testing.T) {
	home := t.TempDir()
	grokHome := t.TempDir()
	t.Setenv("GROK_HOME", grokHome)
	t.Setenv("HOME", t.TempDir())

	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	const sessionID = "seatalk-ok-1"
	const runnerID = "01a01e5d-a7eb-7d31-b013-04f57284959f"
	if err := store.CreateSession(sessionID, agentstorage.SessionMeta{
		SessionID:       sessionID,
		Runner:          "grok-tty",
		RunnerSessionID: runnerID,
		Status:          "exited",
		Workspace:       "/tmp/ws",
	}); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(grokHome, "sessions", "%2Ftmp%2Fws", runnerID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), []byte("{}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var resumeN, runN int
	exited := true
	err = AutoSendOrResume(context.Background(), Opts{
		SessionID:   sessionID,
		Prompt:      "follow up",
		AgentRunner: "grok-tty",
		Store:       store,
		Probe: func(store agentstorage.Store, m agentstorage.SessionMeta) (ProbeReport, error) {
			return ProbeReport{ResumeReady: true, RunnerExited: &exited}, nil
		},
		ResumeSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta) error {
			resumeN++
			return nil
		},
		RunSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta, found bool) error {
			runN++
			return fmt.Errorf("RunSession should not run when local grok data exists")
		},
	})
	if err != nil {
		t.Fatalf("AutoSendOrResume: %v", err)
	}
	if resumeN != 1 || runN != 0 {
		t.Fatalf("resumeN=%d runN=%d; want resume=1 run=0", resumeN, runN)
	}
}
