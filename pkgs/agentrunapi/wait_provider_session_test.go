package agentrunapi

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
)

func TestPickResolvePIDs(t *testing.T) {
	got := pickResolvePIDs(10, 20)
	if len(got) != 2 || got[0] != 20 || got[1] != 10 {
		t.Fatalf("prefer command then serve: %v", got)
	}
	got = pickResolvePIDs(10, 0)
	if len(got) != 1 || got[0] != 10 {
		t.Fatalf("serve only: %v", got)
	}
	got = pickResolvePIDs(5, 5)
	if len(got) != 1 || got[0] != 5 {
		t.Fatalf("dedupe: %v", got)
	}
}

func TestProviderKindForRunner(t *testing.T) {
	if got := providerKindForRunner("grok-tty"); got != "grok" {
		t.Fatalf("got %q", got)
	}
	if got := providerKindForRunner("codex-tty"); got != "codex" {
		t.Fatalf("got %q", got)
	}
}

func TestWaitProviderSessionID_successOnCommandPID(t *testing.T) {
	var sleeps int
	res, err := WaitProviderSessionID(WaitProviderSessionOpts{
		Home:      "/tmp/agent-run-home",
		Runner:    "grok-tty",
		SessionID: "sess-1",
		Timeout:   time.Second,
		ReadRegistryPIDs: func(home, runner, sessionID string) (int, int, error) {
			if home == "" || runner == "" || sessionID == "" {
				t.Fatalf("missing args")
			}
			return 100, 200, nil
		},
		ResolveFromPID: func(pid int) (*procresolve.Result, error) {
			if pid != 200 && pid != 100 {
				return nil, fmt.Errorf("unexpected pid %d", pid)
			}
			if pid == 200 {
				return &procresolve.Result{
					Kind:       "grok",
					SessionID:  "01a064d2-70ec-7162-b36b-8a50ba323569",
					Confidence: "hard",
					RunnerPID:  200,
				}, nil
			}
			return &procresolve.Result{Kind: "none"}, nil
		},
		Sleep: func(time.Duration) error {
			sleeps++
			return nil
		},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.ProviderSessionID != "01a064d2-70ec-7162-b36b-8a50ba323569" {
		t.Fatalf("id=%q", res.ProviderSessionID)
	}
	if res.Kind != "grok" || res.CommandPID != 200 || res.ServePID != 100 {
		t.Fatalf("res=%+v", res)
	}
	if sleeps != 0 {
		t.Fatalf("should succeed first poll, sleeps=%d", sleeps)
	}
}

func TestWaitProviderSessionID_pollsUntilReady(t *testing.T) {
	n := 0
	res, err := WaitProviderSessionID(WaitProviderSessionOpts{
		Home:      "/tmp/h",
		Runner:    "codex-tty",
		SessionID: "s2",
		Timeout:   time.Second,
		PollInterval: time.Millisecond,
		ReadRegistryPIDs: func(home, runner, sessionID string) (int, int, error) {
			n++
			if n < 3 {
				return 0, 0, fmt.Errorf("not ready")
			}
			return 1, 2, nil
		},
		ResolveFromPID: func(pid int) (*procresolve.Result, error) {
			return &procresolve.Result{
				Kind:       "codex",
				SessionID:  "019f283a-aaaa-7bbb-cccc-dddddddddddd",
				Confidence: "hard",
				RunnerPID:  pid,
			}, nil
		},
		Sleep: func(time.Duration) error { return nil },
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Kind != "codex" || res.ProviderSessionID == "" {
		t.Fatalf("res=%+v", res)
	}
	if n < 3 {
		t.Fatalf("expected polls, n=%d", n)
	}
}

func TestWaitProviderSessionID_timeout(t *testing.T) {
	_, err := WaitProviderSessionID(WaitProviderSessionOpts{
		Home:      "/tmp/h",
		Runner:    "grok-tty",
		SessionID: "s3",
		Timeout:   5 * time.Millisecond,
		PollInterval: time.Millisecond,
		ReadRegistryPIDs: func(home, runner, sessionID string) (int, int, error) {
			return 0, 0, fmt.Errorf("missing")
		},
		ResolveFromPID: func(pid int) (*procresolve.Result, error) {
			return &procresolve.Result{Kind: "none"}, nil
		},
		Sleep: func(time.Duration) error { return nil },
	})
	if err == nil {
		t.Fatal("want timeout")
	}
	if got := err.Error(); !strings.Contains(got, "timeout") || !strings.Contains(got, "s3") {
		t.Fatalf("err=%q", got)
	}
}

func TestWaitProviderSessionID_validation(t *testing.T) {
	if _, err := WaitProviderSessionID(WaitProviderSessionOpts{}); err == nil {
		t.Fatal("want home required")
	}
}
