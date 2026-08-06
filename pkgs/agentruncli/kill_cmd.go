package agentruncli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

const killHelp = `
Usage: agent-run kill [OPTIONS] <session-id>

Stop a live TTY session by registry id (SIGTERM → wait → SIGKILL, remove registry).

Options:
  --dry-run    report what would be stopped without terminating
  -h, --help   show help
`

// runKill implements top-level `agent-run kill` and `agent-run tty kill`.
func runKill(args []string) error {
	var dryRun bool
	remaining, err := flags.Bool("--dry-run", &dryRun).
		Help("-h,--help", killHelp).
		HelpNoExit().
		Parse(args)
	if err == flags.ErrHelp {
		return nil
	}
	if err != nil {
		return err
	}
	if len(remaining) == 0 {
		return fmt.Errorf("kill: requires <session-id>")
	}
	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("kill: requires <session-id>")
	}

	store, err := openStore()
	if err != nil {
		return err
	}
	home := store.Home()

	// Resolve registry without LookupSession TCP pruning — reclaim reads
	// the entry by id even when listen_addr is unreachable (test fixtures).
	ttySess, err := agenttty.ResolveByTerminalID(home, sessionID)
	if err != nil {
		if sessionKnownNotRunning(store, home, sessionID) {
			fmt.Fprintf(os.Stderr, "warning: session %s not running\n", sessionID)
			return nil
		}
		// Match sealed unknown-session wording (not found / expired).
		return fmt.Errorf("session %s not found or expired", sessionID)
	}

	if dryRun {
		fmt.Printf("dry-run: would stop %s\n", sessionID)
		return nil
	}

	pid := ttySess.Registry.PID
	cfg := registryConfigForRunner(home, ttySess.RunnerID)
	if reclaimErr := ttywatch.ReclaimSessionID(cfg, sessionID); reclaimErr != nil {
		return reclaimErr
	}
	// Also clear the id under any other provider subdir that might hold it.
	_ = reclaimAcrossProviders(home, sessionID)

	// Best-effort reap when the killed process is our child (L2 in-process
	// Handle: sleep fixtures become zombies until Wait; kill(0) still succeeds).
	reapIfChild(pid)

	fmt.Printf("stopped %s\n", sessionID)
	return nil
}

// sessionKnownNotRunning reports whether sessionID was a known agent-run
// session (meta present) or is referenced as a terminal id, so a second kill
// can be idempotent with a warning instead of "not found".
func sessionKnownNotRunning(store agentstorage.Store, home, sessionID string) bool {
	if _, err := store.GetSession(sessionID); err == nil {
		return true
	}
	// Match by terminal_session_id when agent id differs from registry id.
	sessions, err := store.ListSessions()
	if err != nil {
		return false
	}
	for _, meta := range sessions {
		if strings.TrimSpace(meta.TerminalSessionID) == sessionID {
			return true
		}
		if strings.TrimSpace(meta.SessionID) == sessionID {
			return true
		}
	}
	_ = home
	return false
}

// reapIfChild waits briefly on pid so a zombie child is reaped. No-op for
// non-children (waitpid returns ECHILD immediately) and times out if still live.
func reapIfChild(pid int) {
	if pid <= 0 {
		return
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	done := make(chan struct{})
	go func() {
		_, _ = p.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
	}
}
