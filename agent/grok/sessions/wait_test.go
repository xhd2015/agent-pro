package sessions

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"
)

func TestWait_outsideTurnImmediate(t *testing.T) {
	home, sid := writeWaitFixture(t, []string{
		`{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"hi"}}`,
		`{"sessionUpdate":"turn_completed","prompt_id":"p1","stop_reason":"end_turn"}`,
	})
	live := runningLive(sid)
	res, err := Wait(home, sid, WaitOpts{Live: live, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if res.Reason != WaitReasonTurnCompleted {
		t.Fatalf("reason=%q", res.Reason)
	}
}

func TestWait_notRunning(t *testing.T) {
	home, sid := writeWaitFixture(t, nil)
	_, err := Wait(home, sid, WaitOpts{
		Live: &LiveOptions{
			ListProcs: func() []procresolve.Proc { return nil },
			Lsof:      func(int) []string { return nil },
		},
		Timeout: time.Second,
	})
	if err == nil || !strings.Contains(err.Error(), "not running") {
		t.Fatalf("err=%v", err)
	}
}

func TestWait_inProgressThenCompleted(t *testing.T) {
	home, sid := writeWaitFixture(t, []string{
		`{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"hi"}}`,
		`{"sessionUpdate":"tool_call","toolCallId":"t1"}`,
	})
	live := runningLive(sid)
	res, err := Wait(home, sid, WaitOpts{
		Live:    live,
		Timeout: 2 * time.Second,
		WatchLine: func(ctx context.Context, path string, opts logs.WatchLineOptions, fn func(line string) error) error {
			line := `{"sessionUpdate":"turn_completed","prompt_id":"p2","stop_reason":"end_turn"}`
			_ = fn(line)
			<-ctx.Done()
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Reason != WaitReasonTurnCompleted {
		t.Fatalf("reason=%q", res.Reason)
	}
}

func TestWait_reclassifyCatchesEOFRace(t *testing.T) {
	// turn_completed lands after Phase A but before WatchLine (EOF start) sees it.
	home, sid := writeWaitFixture(t, []string{
		`{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"hi"}}`,
	})
	live := runningLive(sid)
	dir := filepath.Join(home, "sessions", url.PathEscape("/tmp/kck-wait-proj"), sid)
	updates := filepath.Join(dir, "updates.jsonl")
	res, err := Wait(home, sid, WaitOpts{
		Live:           live,
		Timeout:        2 * time.Second,
		StatusInterval: 30 * time.Millisecond,
		WatchLine: func(ctx context.Context, path string, opts logs.WatchLineOptions, fn func(line string) error) error {
			f, err := os.OpenFile(updates, os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				return err
			}
			_, _ = f.WriteString(`{"sessionUpdate":"turn_completed","prompt_id":"p3","stop_reason":"end_turn"}` + "\n")
			_ = f.Close()
			<-ctx.Done()
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Reason != WaitReasonTurnCompleted {
		t.Fatalf("reason=%q", res.Reason)
	}
}

func TestWait_timeout(t *testing.T) {
	home, sid := writeWaitFixture(t, []string{
		`{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"hi"}}`,
	})
	live := runningLive(sid)
	_, err := Wait(home, sid, WaitOpts{
		Live:           live,
		Timeout:        50 * time.Millisecond,
		StatusInterval: 20 * time.Millisecond,
		WatchLine: func(ctx context.Context, path string, opts logs.WatchLineOptions, fn func(line string) error) error {
			<-ctx.Done()
			return nil
		},
	})
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("err=%v", err)
	}
}

func TestSessionUpdateKind_envelope(t *testing.T) {
	line := `{"method":"session/update","params":{"update":{"sessionUpdate":"turn_completed","prompt_id":"x"}}}`
	kind, ok := sessionUpdateKind(line)
	if !ok || kind != "turn_completed" {
		t.Fatalf("got %q ok=%v", kind, ok)
	}
}

func writeWaitFixture(t *testing.T, updateLines []string) (grokHome, sessionID string) {
	t.Helper()
	sessionID = "01a06699-aaaa-7aaa-aaaa-aaaaaaaaaa01"
	grokHome = t.TempDir()
	cwd := "/tmp/kck-wait-proj"
	dir := filepath.Join(grokHome, "sessions", url.PathEscape(cwd), sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	summary := map[string]any{
		"info":           map[string]any{"id": sessionID, "cwd": cwd},
		"created_at":     "2026-07-01T10:00:00.000Z",
		"updated_at":     "2026-07-01T11:00:00.000Z",
		"last_active_at": "2026-07-01T11:00:00.000Z",
	}
	body, _ := json.Marshal(summary)
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), append(body, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	if updateLines != nil {
		var b []byte
		for _, line := range updateLines {
			b = append(b, line...)
			b = append(b, '\n')
		}
		if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return grokHome, sessionID
}

func runningLive(sessionID string) *LiveOptions {
	open := "/Users/fixture/.grok/sessions/%2Ftmp%2Fproj/" + sessionID + "/updates.jsonl"
	return &LiveOptions{
		ListProcs: func() []procresolve.Proc {
			return []procresolve.Proc{{PID: 7001, PPID: 1, Cmd: "/usr/local/bin/grok"}}
		},
		Lsof: func(pid int) []string {
			if pid == 7001 {
				return []string{open}
			}
			return nil
		},
	}
}

