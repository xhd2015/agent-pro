package ttyrunner

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/groktty"
)

// LookupSession searches registered providers' registry dirs in registration order
// and returns the first reachable live entry. Stale unreachable entries are removed.
func LookupSession(home, terminalSessionID string) (*RegistryEntry, string, error) {
	ensureStubRegistered()
	var staleErrs []string
	for _, p := range ProviderListSorted() {
		entry, err := groktty.ReadRegistryFor(home, p.RegistryDir, terminalSessionID, p.ID)
		if err == nil {
			if registryEntryReachable(entry) {
				return entry, p.ID, nil
			}
			staleErrs = append(staleErrs, fmt.Sprintf("%s %s unreachable", p.ID, entry.ListenAddr))
			groktty.RemoveRegistryFor(home, p.RegistryDir, terminalSessionID)
			continue
		}
		if !os.IsNotExist(err) && !strings.Contains(err.Error(), "not found or expired") {
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
	reachable := registryEntryReachable(entry)
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
		scrollback := fetchScrollbackSnapshot(entry.ListenAddr, terminalSessionID)
		if live := provider.DetectScreenStatus(scrollback); live != "" && live != "unknown" {
			sess.ScreenStatus = live
		}
	}
	return sess, nil
}

func findRegistryEntry(home, terminalSessionID string) (*RegistryEntry, string, error) {
	ensureStubRegistered()
	for _, p := range ProviderListSorted() {
		entry, err := groktty.ReadRegistryFor(home, p.RegistryDir, terminalSessionID, p.ID)
		if err == nil {
			return entry, p.ID, nil
		}
		if !os.IsNotExist(err) && !strings.Contains(err.Error(), "not found or expired") {
			return nil, "", err
		}
	}
	return nil, "", fmt.Errorf("tty session not found or expired")
}

// ResolveByAgentSession resolves via meta.terminal_session_id then live registry.
func ResolveByAgentSession(store agentstorage.Store, runner, agentSessionID string) (*TTYSession, error) {
	ensureStubRegistered()
	sess, err := store.GetSession(runner, agentSessionID)
	if err != nil {
		return nil, err
	}
	terminalSessionID := strings.TrimSpace(sess.Meta.TerminalSessionID)
	if terminalSessionID == "" {
		return nil, fmt.Errorf("tty session not found or expired")
	}
	entry, err := groktty.ReadRegistryFor(store.Home(), runner+"-registry", terminalSessionID, runner)
	if err != nil {
		if snap := readTTYJSON(store.Home(), runner, agentSessionID); snap != nil {
			return nil, fmt.Errorf("tty session not found or expired")
		}
		return nil, err
	}
	if !registryEntryReachable(entry) {
		groktty.RemoveRegistryFor(store.Home(), runner+"-registry", terminalSessionID)
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
		scrollback := fetchScrollbackSnapshot(entry.ListenAddr, terminalSessionID)
		result.ScreenStatus = provider.DetectScreenStatus(scrollback)
	}
	return result, nil
}

func registryEntryReachable(entry *RegistryEntry) bool {
	if entry == nil || strings.TrimSpace(entry.ListenAddr) == "" {
		return false
	}
	conn, err := net.DialTimeout("tcp", entry.ListenAddr, 300*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func ttyJSONPath(home, runner, agentSessionID string) string {
	return filepath.Join(home, "sessions", runner, agentSessionID, "tty.json")
}

// WriteTTYJSON dual-writes the denormalized TTY snapshot.
func WriteTTYJSON(home string, snap TTYSnapshot) error {
	if snap.AgentSessionID == "" {
		return nil
	}
	dir := filepath.Dir(ttyJSONPath(home, snap.RunnerID, snap.AgentSessionID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if snap.CreatedAt == "" {
		snap.CreatedAt = nowRFC3339()
	}
	data, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	return os.WriteFile(ttyJSONPath(home, snap.RunnerID, snap.AgentSessionID), data, 0644)
}

// MarkTTYDead sets alive=false on tty.json (best-effort).
func MarkTTYDead(home, runner, agentSessionID string) {
	snap := readTTYJSON(home, runner, agentSessionID)
	if snap == nil {
		return
	}
	snap.Alive = false
	_ = WriteTTYJSON(home, *snap)
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
	sessionsDir := filepath.Join(home, "sessions", runner)
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, ""
	}
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		agentID := ent.Name()
		snap := readTTYJSON(home, runner, agentID)
		if snap != nil && snap.TerminalSessionID == terminalSessionID {
			return snap, agentID
		}
	}
	return nil, ""
}

func findAgentSessionByTerminalID(home, runner, terminalSessionID string) string {
	_, agentID := findTTYJSONForTerminal(home, runner, terminalSessionID)
	if agentID != "" {
		return agentID
	}
	sessionsDir := filepath.Join(home, "sessions", runner)
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return ""
	}
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		metaPath := filepath.Join(sessionsDir, ent.Name(), "meta.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}
		var meta struct {
			TerminalSessionID string `json:"terminal_session_id"`
		}
		if json.Unmarshal(data, &meta) == nil && meta.TerminalSessionID == terminalSessionID {
			return ent.Name()
		}
	}
	return ""
}