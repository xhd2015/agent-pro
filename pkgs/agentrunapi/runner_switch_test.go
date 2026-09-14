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

func writeLocalGrokSession(t *testing.T, grokHome, runnerID string) {
	t.Helper()
	dir := filepath.Join(grokHome, "sessions", "%2Ftmp%2Fws", runnerID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), []byte("{}\n"), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestAutoSendOrResume_switchToNewFamilyModeRunsKeepsOldBind(t *testing.T) {
	home := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	const sessionID = "seatalk-switch-1"
	const codexID = "019fdca1-3893-7fa3-a8aa-ebc1ccc750a0"
	if err := store.CreateSession(sessionID, agentstorage.SessionMeta{
		SessionID:       sessionID,
		Runner:          "codex-tty",
		RunnerSessionID: codexID,
		Status:          "exited",
	}); err != nil {
		t.Fatal(err)
	}

	var resumeN, runN, sendN int
	var runMeta agentstorage.SessionMeta
	exited := true
	var stderr bytes.Buffer
	err = AutoSendOrResume(context.Background(), Opts{
		SessionID:   sessionID,
		Prompt:      "follow up",
		AgentRunner: "grok-tty",
		Store:       store,
		Stderr:      &stderr,
		Probe: func(store agentstorage.Store, m agentstorage.SessionMeta) (ProbeReport, error) {
			if strings.TrimSpace(m.RunnerSessionID) == "" {
				return ProbeReport{}, nil
			}
			return ProbeReport{ResumeReady: true, RunnerExited: &exited}, nil
		},
		ResumeSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta) error {
			resumeN++
			return fmt.Errorf("resume should not run on first grok visit")
		},
		SendLive: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta) error {
			sendN++
			return fmt.Errorf("send should not run on first grok visit")
		},
		RunSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta, found bool) error {
			runN++
			runMeta = m
			if !found {
				t.Fatal("expected found=true")
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("AutoSendOrResume: %v", err)
	}
	if resumeN != 0 || sendN != 0 || runN != 1 {
		t.Fatalf("resumeN=%d sendN=%d runN=%d", resumeN, sendN, runN)
	}
	if strings.TrimSpace(runMeta.RunnerSessionID) != "" {
		t.Fatalf("ModeRun live slot should be empty, got %q", runMeta.RunnerSessionID)
	}
	sess, err := store.GetSession(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if sess.Meta.Runner != "grok-tty" {
		t.Fatalf("runner=%q", sess.Meta.Runner)
	}
	if sess.Meta.RunnerSessions[agentstorage.RunnerFamilyCodex] != codexID {
		t.Fatalf("codex bind lost: %+v", sess.Meta.RunnerSessions)
	}
	if !strings.Contains(stderr.String(), "starting fresh") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestAutoSendOrResume_switchBackResumesMappedFamily(t *testing.T) {
	home := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	const sessionID = "seatalk-switch-back"
	const codexID = "019fdca1-3893-7fa3-a8aa-ebc1ccc750a0"
	const grokID = "01a01e5d-a7eb-7d31-b013-04f57284959f"
	if err := store.CreateSession(sessionID, agentstorage.SessionMeta{
		SessionID:       sessionID,
		Runner:          "grok-tty",
		RunnerSessionID: grokID,
		RunnerSessions: map[string]string{
			agentstorage.RunnerFamilyGrok:  grokID,
			agentstorage.RunnerFamilyCodex: codexID,
		},
		Status: "exited",
	}); err != nil {
		t.Fatal(err)
	}

	var resumeN, runN int
	var resumeMeta agentstorage.SessionMeta
	exited := true
	var stderr bytes.Buffer
	err = AutoSendOrResume(context.Background(), Opts{
		SessionID:   sessionID,
		Prompt:      "hello again",
		AgentRunner: "codex-tty",
		Store:       store,
		Stderr:      &stderr,
		Probe: func(store agentstorage.Store, m agentstorage.SessionMeta) (ProbeReport, error) {
			if strings.TrimSpace(m.RunnerSessionID) == "" {
				return ProbeReport{}, nil
			}
			return ProbeReport{ResumeReady: true, RunnerExited: &exited}, nil
		},
		ResumeSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta) error {
			resumeN++
			resumeMeta = m
			return nil
		},
		RunSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta, found bool) error {
			runN++
			return fmt.Errorf("RunSession should not run when mapped codex bind exists")
		},
	})
	if err != nil {
		t.Fatalf("AutoSendOrResume: %v", err)
	}
	if resumeN != 1 || runN != 0 {
		t.Fatalf("resumeN=%d runN=%d", resumeN, runN)
	}
	if resumeMeta.RunnerSessionID != codexID {
		t.Fatalf("resumed id=%q want %q", resumeMeta.RunnerSessionID, codexID)
	}
	sess, err := store.GetSession(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if sess.Meta.RunnerSessions[agentstorage.RunnerFamilyGrok] != grokID {
		t.Fatalf("grok bind lost: %+v", sess.Meta.RunnerSessions)
	}
	if !strings.Contains(stderr.String(), "resuming") || !strings.Contains(stderr.String(), codexID) {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestAutoSendOrResume_liveOtherFamilyDoesNotSend(t *testing.T) {
	home := t.TempDir()
	grokHome := t.TempDir()
	t.Setenv("GROK_HOME", grokHome)
	t.Setenv("HOME", t.TempDir())
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	const sessionID = "seatalk-live-switch"
	const grokID = "01a01e5d-a7eb-7d31-b013-04f57284959f"
	writeLocalGrokSession(t, grokHome, grokID)
	live := false
	if err := store.CreateSession(sessionID, agentstorage.SessionMeta{
		SessionID:       sessionID,
		Runner:          "codex-tty",
		RunnerSessionID: "codex-live",
		RunnerSessions: map[string]string{
			agentstorage.RunnerFamilyCodex: "codex-live",
			agentstorage.RunnerFamilyGrok:  grokID,
		},
		Status:    "running",
		Workspace: "/tmp/ws",
	}); err != nil {
		t.Fatal(err)
	}

	var sendN, resumeN, runN int
	err = AutoSendOrResume(context.Background(), Opts{
		SessionID:   sessionID,
		Prompt:      "now on grok",
		AgentRunner: "grok-tty",
		Store:       store,
		Probe: func(store agentstorage.Store, m agentstorage.SessionMeta) (ProbeReport, error) {
			// Pretend Classify still sees a live TTY (old Codex window).
			return ProbeReport{RunnerExited: &live}, nil
		},
		SendLive: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta) error {
			sendN++
			return fmt.Errorf("must not inject into the other family's live TTY")
		},
		ResumeSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta) error {
			resumeN++
			if m.RunnerSessionID != grokID {
				t.Fatalf("resume id=%q want %q", m.RunnerSessionID, grokID)
			}
			return nil
		},
		RunSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta, found bool) error {
			runN++
			return fmt.Errorf("should resume mapped grok, not ModeRun")
		},
	})
	if err != nil {
		t.Fatalf("AutoSendOrResume: %v", err)
	}
	if sendN != 0 || resumeN != 1 || runN != 0 {
		t.Fatalf("sendN=%d resumeN=%d runN=%d", sendN, resumeN, runN)
	}
}

func TestAutoSendOrResume_sameFamilyAliasStillResumes(t *testing.T) {
	home := t.TempDir()
	grokHome := t.TempDir()
	t.Setenv("GROK_HOME", grokHome)
	t.Setenv("HOME", t.TempDir())
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	const sessionID = "seatalk-alias"
	const grokID = "01a01e5d-a7eb-7d31-b013-04f57284959f"
	writeLocalGrokSession(t, grokHome, grokID)
	exited := true
	if err := store.CreateSession(sessionID, agentstorage.SessionMeta{
		SessionID:       sessionID,
		Runner:          "grok",
		RunnerSessionID: grokID,
		Status:          "exited",
		Workspace:       "/tmp/ws",
	}); err != nil {
		t.Fatal(err)
	}

	var resumeN, runN int
	err = AutoSendOrResume(context.Background(), Opts{
		SessionID:   sessionID,
		Prompt:      "same family",
		AgentRunner: "grok-tty",
		Store:       store,
		Probe: func(store agentstorage.Store, m agentstorage.SessionMeta) (ProbeReport, error) {
			return ProbeReport{ResumeReady: true, RunnerExited: &exited}, nil
		},
		ResumeSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta) error {
			resumeN++
			if m.RunnerSessionID != grokID {
				t.Fatalf("id=%q", m.RunnerSessionID)
			}
			return nil
		},
		RunSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta, found bool) error {
			runN++
			return fmt.Errorf("alias switch must not ModeRun")
		},
	})
	if err != nil {
		t.Fatalf("AutoSendOrResume: %v", err)
	}
	if resumeN != 1 || runN != 0 {
		t.Fatalf("resumeN=%d runN=%d", resumeN, runN)
	}
}

func TestAutoSendOrResume_emptyAgentRunnerDoesNotRetarget(t *testing.T) {
	home := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	const sessionID = "seatalk-empty-runner"
	const codexID = "019fdca1-3893-7fa3-a8aa-ebc1ccc750a0"
	exited := true
	if err := store.CreateSession(sessionID, agentstorage.SessionMeta{
		SessionID:       sessionID,
		Runner:          "codex-tty",
		RunnerSessionID: codexID,
		Status:          "exited",
	}); err != nil {
		t.Fatal(err)
	}

	var resumeN int
	err = AutoSendOrResume(context.Background(), Opts{
		SessionID: sessionID,
		Prompt:    "keep codex",
		Store:     store,
		Probe: func(store agentstorage.Store, m agentstorage.SessionMeta) (ProbeReport, error) {
			return ProbeReport{ResumeReady: true, RunnerExited: &exited}, nil
		},
		ResumeSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta) error {
			resumeN++
			if m.RunnerSessionID != codexID {
				t.Fatalf("id=%q", m.RunnerSessionID)
			}
			return nil
		},
		RunSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta, found bool) error {
			return fmt.Errorf("should resume")
		},
	})
	if err != nil {
		t.Fatalf("AutoSendOrResume: %v", err)
	}
	if resumeN != 1 {
		t.Fatalf("resumeN=%d", resumeN)
	}
	sess, err := store.GetSession(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if sess.Meta.Runner != "codex-tty" {
		t.Fatalf("runner=%q", sess.Meta.Runner)
	}
}

func TestAutoSendOrResume_staleGrokClearKeepsCodexMap(t *testing.T) {
	home := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	const sessionID = "seatalk-stale-grok"
	const grokID = "01a01e5d-dead-beef-b013-04f57284959f"
	const codexID = "019fdca1-3893-7fa3-a8aa-ebc1ccc750a0"
	exited := true
	if err := store.CreateSession(sessionID, agentstorage.SessionMeta{
		SessionID:       sessionID,
		Runner:          "grok-tty",
		RunnerSessionID: grokID,
		RunnerSessions: map[string]string{
			agentstorage.RunnerFamilyGrok:  grokID,
			agentstorage.RunnerFamilyCodex: codexID,
		},
		Status:    "exited",
		Workspace: "/tmp/ws",
	}); err != nil {
		t.Fatal(err)
	}

	var runN, resumeN int
	t.Setenv("GROK_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
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
			return fmt.Errorf("stale grok must not resume")
		},
		RunSession: func(ctx context.Context, opts Opts, m agentstorage.SessionMeta, found bool) error {
			runN++
			if strings.TrimSpace(m.RunnerSessionID) != "" {
				t.Fatalf("ModeRun live=%q", m.RunnerSessionID)
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("AutoSendOrResume: %v", err)
	}
	if resumeN != 0 || runN != 1 {
		t.Fatalf("resumeN=%d runN=%d", resumeN, runN)
	}
	sess, err := store.GetSession(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if sess.Meta.RunnerSessions[agentstorage.RunnerFamilyCodex] != codexID {
		t.Fatalf("codex bind lost: %+v", sess.Meta.RunnerSessions)
	}
	if _, ok := sess.Meta.RunnerSessions[agentstorage.RunnerFamilyGrok]; ok {
		t.Fatalf("grok slot should be cleared: %+v", sess.Meta.RunnerSessions)
	}
}
