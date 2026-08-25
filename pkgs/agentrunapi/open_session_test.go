package agentrunapi

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func TestIsFocusMiss(t *testing.T) {
	misses := []string{
		"no iTerm candidates found for session x",
		"registry entry not found: /tmp/codex-tty-registry/x.json",
		"no candidates: registry pid missing for terminal session x",
	}
	for _, msg := range misses {
		if !IsFocusMiss(fmt.Errorf("%s", msg)) {
			t.Fatalf("expected miss for %q", msg)
		}
	}
	if IsFocusMiss(fmt.Errorf("multiple iTerm candidates (2); specify --index N")) {
		t.Fatal("ambiguous multi-match must not be a soft miss")
	}
	if IsFocusMiss(nil) {
		t.Fatal("nil must not be a miss")
	}
}

func TestOpenSession_FocusHit(t *testing.T) {
	home := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	const sid = "sess-focus-hit"
	if err := store.CreateSession(sid, agentstorage.SessionMeta{
		SessionID: sid,
		Runner:    "codex-tty",
		Workspace: t.TempDir(),
	}); err != nil {
		t.Fatal(err)
	}

	res, err := OpenSession(OpenSessionOpts{
		Store:     store,
		SessionID: sid,
		FocusSession: func(opts FocusOpts) (FocusCandidate, error) {
			return FocusCandidate{}, nil
		},
		OpenInNewTerminal: func(opts OpenInNewTerminalOpts) error {
			t.Fatal("must not ForceNew on focus hit")
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != OpenSessionActionFocused || res.Runner != "codex-tty" {
		t.Fatalf("%#v", res)
	}
}

func TestOpenSession_RegistryMissForceNews(t *testing.T) {
	home := t.TempDir()
	ws := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	const sid = "sess-registry-miss"
	if err := store.CreateSession(sid, agentstorage.SessionMeta{
		SessionID:         sid,
		Runner:            "codex-tty",
		RunnerSessionID:   "01a0codex",
		TerminalSessionID: sid,
		Workspace:         ws,
	}); err != nil {
		t.Fatal(err)
	}

	var opened OpenInNewTerminalOpts
	res, err := OpenSession(OpenSessionOpts{
		Store:     store,
		SessionID: sid,
		Driver:    agentdriver.Driver{Binary: "/abs/marcus", Args: []string{"agent-run"}},
		FocusSession: func(opts FocusOpts) (FocusCandidate, error) {
			return FocusCandidate{}, fmt.Errorf("registry entry not found: %s/codex-tty-registry/%s.json", home, sid)
		},
		OpenInNewTerminal: func(opts OpenInNewTerminalOpts) error {
			opened = opts
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != OpenSessionActionResumed {
		t.Fatalf("action=%q", res.Action)
	}
	if opened.WorkspaceDir != ws {
		t.Fatalf("WorkspaceDir=%q", opened.WorkspaceDir)
	}
	fu := opened.FollowUpOpts
	if fu.SessionID != sid || !fu.Open || !fu.AllowRelocateResumeSessionDir {
		t.Fatalf("FollowUpOpts=%#v", fu)
	}
	if fu.Driver.Binary != "/abs/marcus" || len(fu.Driver.Args) == 0 || fu.Driver.Args[0] != "agent-run" {
		t.Fatalf("Driver=%#v", fu.Driver)
	}
}

func TestOpenSession_HardFocusError(t *testing.T) {
	home := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	const sid = "sess-hard"
	if err := store.CreateSession(sid, agentstorage.SessionMeta{
		SessionID: sid,
		Runner:    "codex-tty",
		Workspace: t.TempDir(),
	}); err != nil {
		t.Fatal(err)
	}
	_, err = OpenSession(OpenSessionOpts{
		Store:     store,
		SessionID: sid,
		FocusSession: func(opts FocusOpts) (FocusCandidate, error) {
			return FocusCandidate{}, fmt.Errorf("multiple iTerm candidates (2); specify --index N")
		},
		OpenInNewTerminal: func(opts OpenInNewTerminalOpts) error {
			t.Fatal("must not ForceNew on hard focus error")
			return nil
		},
	})
	if err == nil || !strings.Contains(err.Error(), "multiple iTerm candidates") {
		t.Fatalf("got %v", err)
	}
}

func TestOpenSession_EmptyWorkspaceOnMiss(t *testing.T) {
	home := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	const sid = "sess-no-ws"
	if err := store.CreateSession(sid, agentstorage.SessionMeta{
		SessionID: sid,
		Runner:    "codex-tty",
	}); err != nil {
		t.Fatal(err)
	}
	_, err = OpenSession(OpenSessionOpts{
		Store:     store,
		SessionID: sid,
		FocusSession: func(opts FocusOpts) (FocusCandidate, error) {
			return FocusCandidate{}, fmt.Errorf("no iTerm candidates found for session %s", sid)
		},
	})
	if err == nil || !strings.Contains(err.Error(), "empty workspace") {
		t.Fatalf("got %v", err)
	}
}
