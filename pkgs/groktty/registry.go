package groktty

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const registryDirName = "grok-tty-registry"

// RegistryEntry maps a grok-tty session id to the adhoc ptywrap listen address.
type RegistryEntry struct {
	SessionID  string `json:"session_id"`
	ListenAddr string `json:"listen_addr"`
	PID        int    `json:"pid"`
	CreatedAt  string `json:"created_at"`
}

func registryDir(home string) string {
	return registryDirFor(home, registryDirName)
}

func registryPath(home, sessionID string) string {
	return registryPathFor(home, registryDirName, sessionID)
}

func registryDirFor(home, dirName string) string {
	return filepath.Join(home, dirName)
}

func registryPathFor(home, dirName, sessionID string) string {
	return filepath.Join(registryDirFor(home, dirName), sessionID+".json")
}

// WriteRegistry creates the registry file for a live grok-tty session.
func WriteRegistry(home string, entry RegistryEntry) error {
	return WriteRegistryFor(home, registryDirName, entry)
}

// WriteRegistryFor creates a live TTY registry file in a provider-specific registry dir.
func WriteRegistryFor(home, dirName string, entry RegistryEntry) error {
	if err := os.MkdirAll(registryDirFor(home, dirName), 0755); err != nil {
		return err
	}
	if entry.CreatedAt == "" {
		entry.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(registryPathFor(home, dirName, entry.SessionID), data, 0644)
}

// ReadRegistry loads a grok-tty session registry entry.
func ReadRegistry(home, sessionID string) (*RegistryEntry, error) {
	return ReadRegistryFor(home, registryDirName, sessionID, "grok-tty")
}

// ReadRegistryFor loads a provider-specific TTY session registry entry.
func ReadRegistryFor(home, dirName, sessionID, label string) (*RegistryEntry, error) {
	data, err := os.ReadFile(registryPathFor(home, dirName, sessionID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s session not found or expired", label)
		}
		return nil, err
	}
	var entry RegistryEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	if entry.ListenAddr == "" {
		return nil, fmt.Errorf("%s session not found or expired", label)
	}
	return &entry, nil
}

// ReadSupportedRegistry searches known TTY registry dirs deterministically and
// returns the first reachable live entry.
func ReadSupportedRegistry(home, sessionID string) (*RegistryEntry, string, error) {
	candidates := []struct {
		dir   string
		label string
	}{
		{dir: registryDirName, label: "grok-tty"},
		{dir: "codex-tty-registry", label: "codex-tty"},
	}
	var staleErrs []string
	for _, c := range candidates {
		entry, err := ReadRegistryFor(home, c.dir, sessionID, c.label)
		if err == nil {
			if registryEntryReachable(entry) {
				return entry, c.label, nil
			}
			staleErrs = append(staleErrs, fmt.Sprintf("%s %s unreachable", c.label, entry.ListenAddr))
			RemoveRegistryFor(home, c.dir, sessionID)
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

// RemoveRegistry deletes the registry file for a session.
func RemoveRegistry(home, sessionID string) {
	RemoveRegistryFor(home, registryDirName, sessionID)
}

// RemoveRegistryFor deletes the registry file for a provider-specific session.
func RemoveRegistryFor(home, dirName, sessionID string) {
	_ = os.Remove(registryPathFor(home, dirName, sessionID))
}
