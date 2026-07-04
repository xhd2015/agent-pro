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

// ReserveRegistrySessionID returns the next session-N id under flock.
func ReserveRegistrySessionID(home string) (string, func(), error) {
	dir := registryDir(home)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", nil, err
	}
	lockPath := filepath.Join(dir, ".lock")
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return "", nil, err
	}
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		_ = lockFile.Close()
		return "", nil, err
	}
	release := func() {
		_ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		_ = lockFile.Close()
	}

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
		if !strings.HasPrefix(id, "session-") {
			continue
		}
		entry, err := ReadRegistry(home, id)
		if err != nil {
			continue
		}
		if prune && !tcpReachable(entry.ListenAddr) {
			RemoveRegistry(home, id)
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