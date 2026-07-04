package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"
)

// RegistryEntry maps a tty-watch session id to the embedded ptywrap listen address.
type RegistryEntry struct {
	SessionID  string   `json:"session_id"`
	ListenAddr string   `json:"listen_addr"`
	PID        int      `json:"pid"`
	CreatedAt  string   `json:"created_at"`
	Command    []string `json:"command"`
	Cwd        string   `json:"cwd,omitempty"`
}

// validateSessionID checks custom session id syntax: [a-zA-Z0-9][a-zA-Z0-9._-]*.
func validateSessionID(id string) error {
	if id == "" || strings.HasPrefix(id, ".") {
		return fmt.Errorf(`run: invalid session id %q`, id)
	}
	for i, r := range id {
		if i == 0 {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				return fmt.Errorf(`run: invalid session id %q`, id)
			}
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_' || r == '-' {
			continue
		}
		return fmt.Errorf(`run: invalid session id %q`, id)
	}
	return nil
}

// ReserveCustomSessionID validates id and ensures it is not held by a live session.
// Stale registry entries are pruned so the id can be reused.
func ReserveCustomSessionID(home, sessionID string) (func(), error) {
	if err := validateSessionID(sessionID); err != nil {
		return nil, err
	}
	release, err := acquireRegistryLock(home)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(registryPath(home, sessionID)); err == nil {
		entry, readErr := ReadRegistry(home, sessionID)
		if readErr == nil {
			if tcpReachable(entry.ListenAddr) {
				release()
				return nil, fmt.Errorf(`run: session id %q already in use`, sessionID)
			}
			RemoveRegistryIfMatch(home, sessionID, entry.ListenAddr, entry.PID)
		} else {
			RemoveRegistry(home, sessionID)
		}
	}
	return release, nil
}

func acquireRegistryLock(home string) (func(), error) {
	dir := registryDir(home)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	lockPath := filepath.Join(dir, ".lock")
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		_ = lockFile.Close()
		return nil, err
	}
	return func() {
		_ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		_ = lockFile.Close()
	}, nil
}

// ReserveRegistrySessionID returns the next session-N id under flock.
func ReserveRegistrySessionID(home string) (string, func(), error) {
	release, err := acquireRegistryLock(home)
	if err != nil {
		return "", nil, err
	}
	dir := registryDir(home)

	entries, err := os.ReadDir(dir)
	if err != nil {
		release()
		return "", nil, err
	}
	maxN := 0
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		name := ent.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		if n, ok := registrySessionNumber(id); ok && n > maxN {
			maxN = n
		}
	}
	return fmt.Sprintf("session-%d", maxN+1), release, nil
}

func registrySessionNumber(id string) (int, bool) {
	const prefix = "session-"
	if !strings.HasPrefix(id, prefix) {
		return 0, false
	}
	n, err := strconv.Atoi(id[len(prefix):])
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// WriteRegistry creates the registry file for a live session.
func WriteRegistry(home string, entry RegistryEntry) error {
	if err := os.MkdirAll(registryDir(home), 0755); err != nil {
		return err
	}
	if entry.CreatedAt == "" {
		entry.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(registryPath(home, entry.SessionID), data, 0644)
}

// ReadRegistry loads a session registry entry.
func ReadRegistry(home, sessionID string) (*RegistryEntry, error) {
	data, err := os.ReadFile(registryPath(home, sessionID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("tty-watch session %s not found", sessionID)
		}
		return nil, err
	}
	var entry RegistryEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	if entry.ListenAddr == "" {
		return nil, fmt.Errorf("tty-watch session %s not found", sessionID)
	}
	return &entry, nil
}

// RemoveRegistry deletes the registry file for a session.
func RemoveRegistry(home, sessionID string) {
	_ = os.Remove(registryPath(home, sessionID))
}

// RemoveRegistryIfMatch deletes the registry file only when the on-disk entry still
// belongs to the caller (same listen address and pid). Stale __serve__ cleanup can
// otherwise delete a newer session that reused the same session id.
func RemoveRegistryIfMatch(home, sessionID, listenAddr string, pid int) {
	entry, err := ReadRegistry(home, sessionID)
	if err != nil {
		return
	}
	if entry.ListenAddr != listenAddr || entry.PID != pid {
		return
	}
	RemoveRegistry(home, sessionID)
}

// ListRegistryEntries returns all registry entries, optionally pruning unreachable ones.
func ListRegistryEntries(home string, prune bool) ([]RegistryEntry, error) {
	dir := registryDir(home)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []RegistryEntry
	for _, ent := range entries {
		name := ent.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		entry, err := ReadRegistry(home, id)
		if err != nil {
			continue
		}
		if prune && !tcpReachable(entry.ListenAddr) {
			RemoveRegistryIfMatch(home, id, entry.ListenAddr, entry.PID)
			continue
		}
		out = append(out, *entry)
	}
	return out, nil
}

func tcpReachable(addr string) bool {
	if strings.TrimSpace(addr) == "" {
		return false
	}
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func waitForRegistryEntry(home, sessionID string, timeout time.Duration) (*RegistryEntry, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		entry, err := ReadRegistry(home, sessionID)
		if err == nil && entry.ListenAddr != "" {
			return entry, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, fmt.Errorf("timeout waiting for registry entry %s", sessionID)
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}