package agenttty

import (
	"strings"

	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

// RegistryAgentExited reports whether the PTY agent child has exited while the
// keep-alive serve may still be reachable. Used to fail-fast open/resume waits
// instead of hanging on banner/trust polls + AttachWriter against a dead TUI.
func RegistryAgentExited(home, runnerID, sessionID string) bool {
	home = strings.TrimSpace(home)
	sessionID = strings.TrimSpace(sessionID)
	if home == "" || sessionID == "" {
		return false
	}
	p, ok := Get(runnerID)
	if !ok {
		return false
	}
	subdir := strings.TrimSpace(p.RegistryDir)
	if subdir == "" {
		return false
	}
	entry, err := ttywatch.ReadRegistry(ttywatch.RegistryConfig{Home: home, Subdir: subdir}, sessionID)
	if err != nil || entry == nil {
		return false
	}
	if entry.CommandExited {
		return true
	}
	if entry.CommandPID > 0 && !ttywatch.ProcessAlive(entry.CommandPID) {
		return true
	}
	return false
}
