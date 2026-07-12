package ttywatch

import (
	"os"
	"strings"
	"syscall"
	"time"

	ptyclient "github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/client"
)

// ReclaimSessionID forcefully frees a session id held by a (possibly zombie)
// keep-alive serve so the id can be re-reserved. Best-effort: delete via
// ptywrap when reachable, terminate the serve PID (never self), then remove
// registry + claim files regardless of process teardown success.
//
// No-op when the id is not present. Intended for resume after status reports
// the runner exited while keep-alive still holds the TTY registry id.
func ReclaimSessionID(cfg RegistryConfig, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}

	entry, err := ReadRegistry(cfg, sessionID)
	if err != nil {
		// Drop a live/stale claim even without a registry entry.
		if pid, ok := readSessionClaimPID(cfg, sessionID); ok {
			terminatePID(pid)
			clearSessionClaim(cfg, sessionID)
		}
		return nil
	}

	// Best-effort ptywrap tear-down when the listen endpoint still answers.
	if entry.ListenAddr != "" && tcpReachable(entry.ListenAddr) {
		c := ptyclient.NewClient("http://" + entry.ListenAddr)
		_ = c.Delete(sessionID)
	}

	// Terminate serve PID (zombie keep-alive __serve__). Never kill ourselves.
	if entry.PID > 0 {
		terminatePID(entry.PID)
	}

	// Force-remove so SessionIDInUse cannot still see a reachable stale entry.
	RemoveRegistry(cfg, sessionID)
	return nil
}

func terminatePID(pid int) {
	if pid <= 0 || pid == os.Getpid() || !processAlive(pid) {
		return
	}
	_ = syscall.Kill(pid, syscall.SIGTERM)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && processAlive(pid) {
		time.Sleep(100 * time.Millisecond)
	}
	if processAlive(pid) {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
}
