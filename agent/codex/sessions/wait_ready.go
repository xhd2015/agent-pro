package sessions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"
)

// waitRolloutReady blocks until the Codex rollout JSONL exists for sessionID,
// or errors if the session is abandoned before the rollout appears (terminal
// closed / process tree exit). Races: fsnotify create, blocking flock, and
// kqueue NOTE_EXIT on lock-holder + parents — no PID polling loops.
func waitRolloutReady(ctx context.Context, codexHome, sessionID, lockPath string, opts WaitOpts) (string, error) {
	if path, err := Find(codexHome, sessionID); err == nil {
		return path, nil
	}

	watchCreate := opts.WatchCreateMatch
	if watchCreate == nil {
		watchCreate = logs.WatchCreateMatch
	}
	waitLock := opts.WaitExclusiveLock
	if waitLock == nil {
		waitLock = waitExclusiveLock
	}
	waitExit := opts.WaitProcessesExit
	if waitExit == nil {
		waitExit = waitProcessesExit
	}

	// Orphan lock (no holder): flock succeeds immediately → abandoned if still no file.
	if err := tryExclusiveLockNB(lockPath); err == nil {
		if path, ferr := Find(codexHome, sessionID); ferr == nil {
			return path, nil
		}
		return "", fmt.Errorf("session never created: %s (closed before rollout appeared)", sessionID)
	}

	readyCh := make(chan string, 1)
	abandonCh := make(chan struct{}, 1)
	errCh := make(chan error, 3)

	sessionsRoot := filepath.Join(codexHome, "sessions")
	want := strings.ToLower(sessionID)

	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()

	signalAbandon := func() {
		select {
		case abandonCh <- struct{}{}:
		default:
		}
		watchCancel()
	}

	go func() {
		err := watchCreate(watchCtx, sessionsRoot, logs.WatchCreateMatchOptions{MaxDepth: 4},
			func(path string) bool {
				base := strings.ToLower(filepath.Base(path))
				return strings.Contains(base, "rollout-") && strings.Contains(base, want) && strings.HasSuffix(base, ".jsonl")
			},
			func(path string) error {
				select {
				case readyCh <- path:
				default:
				}
				watchCancel()
				return nil
			},
		)
		if err != nil && watchCtx.Err() == nil {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	go func() {
		if err := waitLock(watchCtx, lockPath); err != nil {
			if watchCtx.Err() != nil {
				return
			}
			select {
			case errCh <- err:
			default:
			}
			return
		}
		signalAbandon()
	}()

	// ASAP path: process-tree exit (iTerm close kills agent-run run; flock alone
	// can lag if __serve/codex briefly linger).
	pids := []int(nil)
	if opts.ReadinessPIDs != nil {
		pids = opts.ReadinessPIDs(lockPath)
	} else {
		list := procresolve.ListLiveProcs
		if opts.Live != nil && opts.Live.ListProcs != nil {
			list = opts.Live.ListProcs
		}
		pids = readinessWatchPIDs(lockPath, list)
	}
	if len(pids) > 0 {
		go func() {
			if err := waitExit(watchCtx, pids); err != nil {
				if watchCtx.Err() != nil {
					return
				}
				select {
				case errCh <- err:
				default:
				}
				return
			}
			signalAbandon()
		}()
	}

	for {
		select {
		case path := <-readyCh:
			if path == "" {
				if p, err := Find(codexHome, sessionID); err == nil {
					return p, nil
				}
				continue
			}
			return path, nil
		case <-abandonCh:
			// Re-check Find (rollout may have landed as the process exited).
			if path, err := Find(codexHome, sessionID); err == nil {
				return path, nil
			}
			return "", fmt.Errorf("session never created: %s (closed before rollout appeared)", sessionID)
		case err := <-errCh:
			if path, ferr := Find(codexHome, sessionID); ferr == nil {
				return path, nil
			}
			return "", err
		case <-ctx.Done():
			if path, err := Find(codexHome, sessionID); err == nil {
				return path, nil
			}
			return "", fmt.Errorf("timeout waiting for rollout for session %s", sessionID)
		}
	}
}

func threadWriterLockPath(codexHome, sessionID string) string {
	return filepath.Join(codexHome, "thread-writer-locks", sessionID+".lock")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isCodexSessionNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "codex session not found")
}

func tryExclusiveLockNB(lockPath string) error {
	f, err := os.OpenFile(lockPath, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return err
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return nil
}

func waitExclusiveLock(ctx context.Context, lockPath string) error {
	f, err := os.OpenFile(lockPath, os.O_RDWR, 0)
	if err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	}()

	select {
	case <-ctx.Done():
		_ = f.Close() // unblock flock waiter
		select {
		case <-done:
		case <-time.After(200 * time.Millisecond):
		}
		return ctx.Err()
	case err := <-done:
		if err == nil {
			_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		}
		_ = f.Close()
		return err
	}
}
