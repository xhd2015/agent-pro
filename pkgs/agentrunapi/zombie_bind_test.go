package agentrunapi

import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func TestTryBindRunnerSessionFromZombie_noopWhenAlreadyBound(t *testing.T) {
	home := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	sid := "sess-bound-1"
	meta := agentstorage.SessionMeta{
		Runner:            "codex-tty",
		SessionID:         sid,
		Status:            "running",
		TerminalSessionID: sid,
		RunnerSessionID:   "019fdca1-3893-7fa3-a8aa-ebc1ccc750a0",
	}
	if err := store.CreateSession(sid, meta); err != nil {
		t.Fatal(err)
	}
	got := tryBindRunnerSessionFromZombie(store, meta)
	if got.RunnerSessionID != meta.RunnerSessionID {
		t.Fatalf("bound id should stay, got %q", got.RunnerSessionID)
	}
}

func TestTryBindRunnerSessionFromZombie_noopWithoutTTY(t *testing.T) {
	home := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	sid := "sess-nobind-1"
	meta := agentstorage.SessionMeta{
		Runner:            "codex-tty",
		SessionID:         sid,
		Status:            "running",
		TerminalSessionID: sid,
	}
	if err := store.CreateSession(sid, meta); err != nil {
		t.Fatal(err)
	}
	got := tryBindRunnerSessionFromZombie(store, meta)
	if got.RunnerSessionID != "" {
		t.Fatalf("expected no bind without tty, got %q", got.RunnerSessionID)
	}
}

func TestIsCodexRunner(t *testing.T) {
	if !isCodexRunner("codex-tty") || !isCodexRunner("codex") {
		t.Fatal("codex runners")
	}
	if isCodexRunner("grok-tty") {
		t.Fatal("grok is not codex")
	}
}

func TestEffectiveRunner(t *testing.T) {
	if got := effectiveRunner(Opts{AgentRunner: "codex-tty"}, agentstorage.SessionMeta{}); got != "codex-tty" {
		t.Fatalf("opts win: %q", got)
	}
	if got := effectiveRunner(Opts{}, agentstorage.SessionMeta{Runner: "grok-tty"}); got != "grok-tty" {
		t.Fatalf("meta: %q", got)
	}
}
