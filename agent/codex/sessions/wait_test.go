package sessions

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"
)

func TestWait_outsideTurnImmediate(t *testing.T) {
	home, sid, _ := writeWaitRollout(t, []string{
		eventMsgLine(boundaryTaskStarted),
		eventMsgLine(boundaryTaskComplete),
	})
	live := runningCodexLive(sid)
	res, err := Wait(home, sid, WaitOpts{Live: live, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if res.Reason != WaitReasonTurnCompleted {
		t.Fatalf("reason=%q", res.Reason)
	}
}

func TestWait_noBoundaryOutsideTurn(t *testing.T) {
	home, sid, _ := writeWaitRollout(t, nil)
	live := runningCodexLive(sid)
	res, err := Wait(home, sid, WaitOpts{Live: live, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if res.Reason != WaitReasonOutsideTurn {
		t.Fatalf("reason=%q", res.Reason)
	}
}

func TestWait_notRunning(t *testing.T) {
	home, sid, _ := writeWaitRollout(t, nil)
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
	home, sid, _ := writeWaitRollout(t, []string{
		eventMsgLine(boundaryTaskStarted),
		`{"timestamp":"2026-08-01T12:00:01.000Z","type":"event_msg","payload":{"type":"item_completed"}}`,
	})
	live := runningCodexLive(sid)
	res, err := Wait(home, sid, WaitOpts{
		Live:    live,
		Timeout: 2 * time.Second,
		WatchLine: func(ctx context.Context, path string, opts logs.WatchLineOptions, fn func(line string) error) error {
			_ = fn(eventMsgLine(boundaryTaskComplete))
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

func TestWait_inProgressThenAborted(t *testing.T) {
	home, sid, _ := writeWaitRollout(t, []string{
		eventMsgLine(boundaryTaskStarted),
	})
	live := runningCodexLive(sid)
	res, err := Wait(home, sid, WaitOpts{
		Live:    live,
		Timeout: 2 * time.Second,
		WatchLine: func(ctx context.Context, path string, opts logs.WatchLineOptions, fn func(line string) error) error {
			_ = fn(eventMsgLine(boundaryTurnAborted))
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
	home, sid, path := writeWaitRollout(t, []string{
		eventMsgLine(boundaryTaskStarted),
	})
	live := runningCodexLive(sid)
	res, err := Wait(home, sid, WaitOpts{
		Live:           live,
		Timeout:        2 * time.Second,
		StatusInterval: 30 * time.Millisecond,
		WatchLine: func(ctx context.Context, path string, opts logs.WatchLineOptions, fn func(line string) error) error {
			f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				return err
			}
			_, _ = f.WriteString(eventMsgLine(boundaryTaskComplete) + "\n")
			_ = f.Close()
			<-ctx.Done()
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Reason != WaitReasonTurnCompleted {
		t.Fatalf("reason=%q path=%s", res.Reason, path)
	}
}

func TestWait_timeout(t *testing.T) {
	home, sid, _ := writeWaitRollout(t, []string{
		eventMsgLine(boundaryTaskStarted),
	})
	live := runningCodexLive(sid)
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

func TestRolloutBoundaryKind(t *testing.T) {
	kind, ok := rolloutBoundaryKind(eventMsgLine(boundaryTaskComplete))
	if !ok || kind != boundaryTaskComplete {
		t.Fatalf("got %q ok=%v", kind, ok)
	}
	if _, ok := rolloutBoundaryKind(`{"type":"response_item","payload":{"type":"message"}}`); ok {
		t.Fatal("response_item must not be a boundary")
	}
}

func TestWait_pendingThenRolloutReady(t *testing.T) {
	home := t.TempDir()
	sid := "01a06699-eeee-7eee-eeee-eeeeeeeeee01"
	lockDir := filepath.Join(home, "thread-writer-locks")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(lockDir, sid+".lock")
	if err := os.WriteFile(lockPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	// Hold exclusive flock so readiness does not treat the lock as orphaned.
	holder, err := os.OpenFile(lockPath, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(holder.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = syscall.Flock(int(holder.Fd()), syscall.LOCK_UN)
		_ = holder.Close()
	}()

	live := runningCodexLive(sid)
	res, err := Wait(home, sid, WaitOpts{
		Live:    live,
		Timeout: 2 * time.Second,
		WaitExclusiveLock: func(ctx context.Context, path string) error {
			<-ctx.Done()
			return ctx.Err()
		},
		WatchCreateMatch: func(ctx context.Context, rootDir string, opts logs.WatchCreateMatchOptions, match func(string) bool, callback func(string) error) error {
			// Simulate rollout appearing while lock still held.
			_, _, path := writeWaitRolloutAt(t, home, sid, []string{
				eventMsgLine(boundaryTaskComplete),
			})
			return callback(path)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Reason != WaitReasonTurnCompleted {
		t.Fatalf("reason=%q", res.Reason)
	}
}

func TestWait_abandonedBeforeRollout(t *testing.T) {
	home := t.TempDir()
	sid := "01a06699-ffff-7fff-ffff-ffffffffffff01"
	lockDir := filepath.Join(home, "thread-writer-locks")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(lockDir, sid+".lock")
	if err := os.WriteFile(lockPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Wait(home, sid, WaitOpts{
		Live:    runningCodexLive(sid),
		Timeout: time.Second,
		WaitExclusiveLock: func(ctx context.Context, path string) error {
			return nil // lock released / orphan
		},
		WatchCreateMatch: func(ctx context.Context, rootDir string, opts logs.WatchCreateMatchOptions, match func(string) bool, callback func(string) error) error {
			<-ctx.Done()
			return nil
		},
	})
	if err == nil || !strings.Contains(err.Error(), "never created") {
		t.Fatalf("err=%v", err)
	}
}

func TestWait_notFoundWithoutLock(t *testing.T) {
	home := t.TempDir()
	_, err := Wait(home, "01a06699-0000-7000-0000-000000000001", WaitOpts{
		Live:    runningCodexLive("01a06699-0000-7000-0000-000000000001"),
		Timeout: time.Second,
	})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("err=%v", err)
	}
}

func TestWait_processExitAbandonsASAP(t *testing.T) {
	home := t.TempDir()
	sid := "01a06699-bbbb-7bbb-bbbb-bbbbbbbbbb99"
	lockDir := filepath.Join(home, "thread-writer-locks")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(lockDir, sid+".lock")
	if err := os.WriteFile(lockPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	holder, err := os.OpenFile(lockPath, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(holder.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = syscall.Flock(int(holder.Fd()), syscall.LOCK_UN)
		_ = holder.Close()
	}()

	exitCh := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Millisecond)
		close(exitCh)
	}()
	start := time.Now()
	_, err = Wait(home, sid, WaitOpts{
		Live:    runningCodexLive(sid),
		Timeout: 2 * time.Second,
		WaitExclusiveLock: func(ctx context.Context, path string) error {
			<-ctx.Done()
			return ctx.Err()
		},
		WatchCreateMatch: func(ctx context.Context, rootDir string, opts logs.WatchCreateMatchOptions, match func(string) bool, callback func(string) error) error {
			<-ctx.Done()
			return nil
		},
		ReadinessPIDs:     func(string) []int { return []int{12345} },
		WaitProcessesExit: func(ctx context.Context, pids []int) error { <-exitCh; return nil },
	})
	elapsed := time.Since(start)
	if err == nil || !strings.Contains(err.Error(), "never created") {
		t.Fatalf("err=%v", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("abandon too slow: %s", elapsed)
	}
}

func eventMsgLine(payloadType string) string {
	return `{"timestamp":"2026-08-01T12:00:00.000Z","type":"event_msg","payload":{"type":"` + payloadType + `"}}`
}

func writeWaitRollout(t *testing.T, extraLines []string) (codexHome, sessionID, path string) {
	t.Helper()
	sessionID = "01a06699-cccc-7ccc-cccc-cccccccccc01"
	codexHome = t.TempDir()
	return writeWaitRolloutAt(t, codexHome, sessionID, extraLines)
}

func writeWaitRolloutAt(t *testing.T, codexHome, sessionID string, extraLines []string) (string, string, string) {
	t.Helper()
	dir := filepath.Join(codexHome, "sessions", "2026", "08", "01")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "rollout-2026-08-01T12-00-00-"+sessionID+".jsonl")
	var b strings.Builder
	b.WriteString(`{"timestamp":"2026-08-01T12:00:00.000Z","type":"session_meta","payload":{"id":"` + sessionID + `","cwd":"/tmp/proj"}}` + "\n")
	for _, line := range extraLines {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return codexHome, sessionID, path
}

func runningCodexLive(sessionID string) *LiveOptions {
	return &LiveOptions{
		ListProcs: func() []procresolve.Proc {
			return []procresolve.Proc{{PID: 4242, PPID: 1, Cmd: "/usr/local/bin/codex"}}
		},
		Lsof: func(pid int) []string {
			if pid == 4242 {
				return []string{codexRolloutPath(sessionID)}
			}
			return nil
		},
	}
}
