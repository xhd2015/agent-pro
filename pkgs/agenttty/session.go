package agenttty

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

// TTYSession is the unified resolver result for status, attach, and send.
type TTYSession struct {
	RunnerID          string
	AgentSessionID    string
	TerminalSessionID string
	Registry          ttywatch.RegistryEntry
	Meta              *agentstorage.SessionMeta
	TTY               *TTYSnapshot
	TCPReachable      bool
	ScreenStatus      string
}

// ProviderRegistrySubdirs returns registry subdir names for all registered providers.
func ProviderRegistrySubdirs() []string {
	subdirs := make([]string, 0, len(providers))
	for _, p := range ProviderListSorted() {
		subdirs = append(subdirs, p.RegistryDir)
	}
	return subdirs
}

// LookupSession searches registered providers' registry dirs in registration order
// and returns the first reachable live entry. Stale unreachable entries are removed.
func LookupSession(home, terminalSessionID string) (*ttywatch.RegistryEntry, string, error) {
	ensureStubRegistered()
	var staleErrs []string
	for _, p := range ProviderListSorted() {
		cfg := ttywatch.RegistryConfig{Home: home, Subdir: p.RegistryDir}
		entry, err := ttywatch.ReadRegistry(cfg, terminalSessionID)
		if err == nil {
			if ttywatch.TCPReachable(entry.ListenAddr) {
				return entry, p.ID, nil
			}
			staleErrs = append(staleErrs, fmt.Sprintf("%s %s unreachable", p.ID, entry.ListenAddr))
			ttywatch.RemoveRegistryIfMatch(cfg, terminalSessionID, entry.ListenAddr, entry.PID)
			continue
		}
		if !registryNotFound(err) {
			return nil, "", err
		}
	}
	if len(staleErrs) > 0 {
		return nil, "", fmt.Errorf("tty session not found or expired (%s)", strings.Join(staleErrs, "; "))
	}
	return nil, "", fmt.Errorf("tty session not found or expired")
}

// ResolveByTerminalID resolves a terminal-session-id across registered providers.
// Unlike LookupSession, stale entries are retained for status probes (TCPReachable=false).
func ResolveByTerminalID(home, terminalSessionID string) (*TTYSession, error) {
	entry, runnerID, err := findRegistryEntry(home, terminalSessionID)
	if err != nil {
		return nil, err
	}
	provider, _ := Get(runnerID)
	reachable := ttywatch.TCPReachable(entry.ListenAddr)
	sess := &TTYSession{
		RunnerID:          runnerID,
		TerminalSessionID: terminalSessionID,
		Registry:          *entry,
		TCPReachable:      reachable,
	}
	if snap, agentID := findTTYJSONForTerminal(home, runnerID, terminalSessionID); snap != nil {
		sess.TTY = snap
		sess.AgentSessionID = agentID
		if snap.ScreenStatus != "" {
			sess.ScreenStatus = snap.ScreenStatus
		}
	}
	if sess.AgentSessionID == "" {
		sess.AgentSessionID = findAgentSessionByTerminalID(home, runnerID, terminalSessionID)
	}
	if reachable && provider.DetectScreenStatus != nil &&
		(sess.ScreenStatus == "" || sess.ScreenStatus == "unknown") {
		scrollback, err := fetchSnapshotBytes(entry.ListenAddr, terminalSessionID)
		if err == nil {
			if live := provider.DetectScreenStatus(scrollback); live != "" && live != "unknown" {
				sess.ScreenStatus = live
			}
		}
	}
	return sess, nil
}

func findRegistryEntry(home, terminalSessionID string) (*ttywatch.RegistryEntry, string, error) {
	ensureStubRegistered()
	for _, p := range ProviderListSorted() {
		cfg := ttywatch.RegistryConfig{Home: home, Subdir: p.RegistryDir}
		entry, err := ttywatch.ReadRegistry(cfg, terminalSessionID)
		if err == nil {
			return entry, p.ID, nil
		}
		if !registryNotFound(err) {
			return nil, "", err
		}
	}
	return nil, "", fmt.Errorf("tty session not found or expired")
}

// ResolveByAgentSession resolves via meta.terminal_session_id then live registry.
func ResolveByAgentSession(store agentstorage.Store, runner, agentSessionID string) (*TTYSession, error) {
	ensureStubRegistered()
	sess, err := store.GetSession(agentSessionID)
	if err != nil {
		return nil, err
	}
	terminalSessionID := strings.TrimSpace(sess.Meta.TerminalSessionID)
	if terminalSessionID == "" {
		if snap := readTTYJSON(store.Home(), runner, agentSessionID); snap != nil {
			terminalSessionID = strings.TrimSpace(snap.TerminalSessionID)
		}
	}
	if terminalSessionID == "" {
		return nil, fmt.Errorf("tty session not found or expired")
	}
	cfg := ttywatch.RegistryConfig{Home: store.Home(), Subdir: runner + "-registry"}
	entry, err := ttywatch.ReadRegistry(cfg, terminalSessionID)
	if err != nil {
		if snap := readTTYJSON(store.Home(), runner, agentSessionID); snap != nil {
			return nil, fmt.Errorf("tty session not found or expired")
		}
		return nil, err
	}
	if !ttywatch.TCPReachable(entry.ListenAddr) {
		ttywatch.RemoveRegistryIfMatch(cfg, terminalSessionID, entry.ListenAddr, entry.PID)
		return nil, fmt.Errorf("tty session not found or expired")
	}
	provider, _ := Get(runner)
	result := &TTYSession{
		RunnerID:          runner,
		AgentSessionID:    agentSessionID,
		TerminalSessionID: terminalSessionID,
		Registry:          *entry,
		Meta:              &sess.Meta,
		TCPReachable:      true,
	}
	if snap := readTTYJSON(store.Home(), runner, agentSessionID); snap != nil {
		result.TTY = snap
		if snap.ScreenStatus != "" {
			result.ScreenStatus = snap.ScreenStatus
		}
	}
	if result.ScreenStatus == "" && provider.DetectScreenStatus != nil {
		scrollback, err := fetchSnapshotBytes(entry.ListenAddr, terminalSessionID)
		if err == nil {
			result.ScreenStatus = provider.DetectScreenStatus(scrollback)
		}
	}
	return result, nil
}

// ResolveTerminalStatus resolves a live terminal for status probes. It avoids
// removing stale registry entries and scans the runner registry when
// meta.terminal_session_id has not been persisted yet for a running session.
func ResolveTerminalStatus(store agentstorage.Store, runner, agentSessionID string) (*TTYSession, error) {
	if sess, err := ResolveByAgentSession(store, runner, agentSessionID); err == nil {
		return sess, nil
	}

	sess, err := store.GetSession(agentSessionID)
	if err != nil {
		return nil, err
	}
	status := sess.Meta.Status
	if status != "running" && status != "finished" {
		return nil, fmt.Errorf("tty session not found or expired")
	}

	provider, ok := Get(runner)
	if !ok {
		return nil, fmt.Errorf("unknown tty runner: %s", runner)
	}

	cfg := ttywatch.RegistryConfig{Home: store.Home(), Subdir: provider.RegistryDir}
	registryDir := filepath.Join(store.Home(), provider.RegistryDir)
	entries, err := os.ReadDir(registryDir)
	if err != nil {
		return nil, fmt.Errorf("tty session not found or expired")
	}

	var fallback *TTYSession
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".json") {
			continue
		}
		terminalID := strings.TrimSuffix(ent.Name(), ".json")
		entry, err := ttywatch.ReadRegistry(cfg, terminalID)
		if err != nil || !ttywatch.TCPReachable(entry.ListenAddr) {
			continue
		}
		linked := findAgentSessionByTerminalID(store.Home(), runner, terminalID)
		if linked != "" && linked != agentSessionID {
			continue
		}
		candidate := &TTYSession{
			RunnerID:          runner,
			AgentSessionID:    agentSessionID,
			TerminalSessionID: terminalID,
			Registry:          *entry,
			Meta:              &sess.Meta,
			TCPReachable:      true,
		}
		if snap := readTTYJSON(store.Home(), runner, agentSessionID); snap != nil {
			candidate.TTY = snap
			if snap.ScreenStatus != "" {
				candidate.ScreenStatus = snap.ScreenStatus
			}
		}
		if linked == agentSessionID {
			return candidate, nil
		}
		fallback = candidate
	}
	if fallback != nil {
		return fallback, nil
	}
	return nil, fmt.Errorf("tty session not found or expired")
}

func registryNotFound(err error) bool {
	if os.IsNotExist(err) {
		return true
	}
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "not found")
}

func readTTYJSON(home, runner, agentSessionID string) *TTYSnapshot {
	data, err := os.ReadFile(ttyJSONPath(home, runner, agentSessionID))
	if err != nil {
		return nil
	}
	var snap TTYSnapshot
	if json.Unmarshal(data, &snap) != nil {
		return nil
	}
	return &snap
}

func findTTYJSONForTerminal(home, runner, terminalSessionID string) (*TTYSnapshot, string) {
	sessionsDir := filepath.Join(home, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, ""
	}
	for _, ent := range entries {
		if !ent.IsDir() || strings.HasPrefix(ent.Name(), ".") {
			continue
		}
		agentID := ent.Name()
		snap := readTTYJSON(home, runner, agentID)
		if snap != nil && snap.TerminalSessionID == terminalSessionID {
			if runner == "" || snap.RunnerID == "" || snap.RunnerID == runner {
				return snap, agentID
			}
		}
	}
	return nil, ""
}

func findAgentSessionByTerminalID(home, runner, terminalSessionID string) string {
	_, agentID := findTTYJSONForTerminal(home, runner, terminalSessionID)
	if agentID != "" {
		return agentID
	}
	sessionsDir := filepath.Join(home, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return ""
	}
	for _, ent := range entries {
		if !ent.IsDir() || strings.HasPrefix(ent.Name(), ".") {
			continue
		}
		metaPath := filepath.Join(sessionsDir, ent.Name(), "meta.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}
		var meta struct {
			Runner            string `json:"runner"`
			TerminalSessionID string `json:"terminal_session_id"`
		}
		if json.Unmarshal(data, &meta) == nil && meta.TerminalSessionID == terminalSessionID {
			if runner == "" || meta.Runner == "" || meta.Runner == runner {
				return ent.Name()
			}
		}
	}
	return ""
}