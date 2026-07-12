package ttywatch

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"
)

const (
	envTTYWatchHome = "TTY_WATCH_HOME"
	defaultHomeDir  = ".tty-watch"
	defaultSubdir   = "registry"

	// registryLockTimeout bounds exclusive flock acquisition so concurrent
	// run cannot hang forever when registry/.lock is held.
	registryLockTimeout = 1500 * time.Millisecond
	registryLockPoll    = 25 * time.Millisecond
)

// RegistryConfig selects the registry directory under a home path.
type RegistryConfig struct {
	Home   string // AGENT_RUN_HOME or TTY_WATCH_HOME
	Subdir string // "registry" for tty-watch; "grok-tty-registry" for agent-run
}

// DefaultRegistryConfig returns the tty-watch default registry layout.
func DefaultRegistryConfig(home string) RegistryConfig {
	return RegistryConfig{Home: home, Subdir: defaultSubdir}
}

// RegistryDir returns {home}/{subdir}/ for the given config.
func RegistryDir(cfg RegistryConfig) string {
	subdir := cfg.Subdir
	if subdir == "" {
		subdir = defaultSubdir
	}
	return filepath.Join(cfg.Home, subdir)
}

// TTYWatchHome returns the tty-watch data directory.
func TTYWatchHome() (string, error) {
	if v := os.Getenv(envTTYWatchHome); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, defaultHomeDir), nil
}

func registryPath(cfg RegistryConfig, sessionID string) string {
	return filepath.Join(RegistryDir(cfg), sessionID+".json")
}

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
// Stale registry entries are pruned so the id can be reused. On success a side-car
// claim file is written so the exclusive flock can be released before __serve__
// registers (the claim is not a *.json registry entry, so waiters do not treat it
// as a ready session).
func ReserveCustomSessionID(cfg RegistryConfig, sessionID string) (func(), error) {
	if err := validateSessionID(sessionID); err != nil {
		return nil, err
	}
	release, err := acquireRegistryLock(cfg)
	if err != nil {
		return nil, err
	}
	if err := pruneStaleSessionID(cfg, sessionID); err != nil {
		release()
		return nil, err
	}
	if sessionIDInUse(cfg, sessionID) {
		release()
		return nil, fmt.Errorf(`run: session id %q already in use`, sessionID)
	}
	if err := writeSessionClaim(cfg, sessionID); err != nil {
		release()
		return nil, err
	}
	return release, nil
}

func acquireRegistryLock(cfg RegistryConfig) (func(), error) {
	dir := RegistryDir(cfg)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	lockPath := filepath.Join(dir, ".lock")
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(registryLockTimeout)
	for {
		err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if err != syscall.EWOULDBLOCK && err != syscall.EAGAIN {
			_ = lockFile.Close()
			return nil, err
		}
		if time.Now().After(deadline) {
			_ = lockFile.Close()
			return nil, formatRegistryLockBusyError(lockPath, registryLockTimeout)
		}
		time.Sleep(registryLockPoll)
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			_ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
			_ = lockFile.Close()
		})
	}, nil
}

// ReserveRegistrySessionID returns the next session-N id under flock and writes
// a side-car claim so the lock can be released before WaitForRegistryEntry.
func ReserveRegistrySessionID(cfg RegistryConfig) (string, func(), error) {
	release, err := acquireRegistryLock(cfg)
	if err != nil {
		return "", nil, err
	}
	dir := RegistryDir(cfg)

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
		var id string
		switch {
		case strings.HasSuffix(name, ".json"):
			id = strings.TrimSuffix(name, ".json")
		case strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".claim"):
			// ".session-3.claim" → "session-3"
			id = strings.TrimSuffix(strings.TrimPrefix(name, "."), ".claim")
		default:
			continue
		}
		if n, ok := registrySessionNumber(id); ok && n > maxN {
			maxN = n
		}
	}
	sessionID := fmt.Sprintf("session-%d", maxN+1)
	if err := writeSessionClaim(cfg, sessionID); err != nil {
		release()
		return "", nil, err
	}
	return sessionID, release, nil
}

func sessionClaimPath(cfg RegistryConfig, sessionID string) string {
	return filepath.Join(RegistryDir(cfg), "."+sessionID+".claim")
}

// writeSessionClaim records a provisional reservation (PID of the reserving
// process) without creating a *.json registry entry.
func writeSessionClaim(cfg RegistryConfig, sessionID string) error {
	if err := os.MkdirAll(RegistryDir(cfg), 0755); err != nil {
		return err
	}
	return os.WriteFile(sessionClaimPath(cfg, sessionID), []byte(strconv.Itoa(os.Getpid())+"\n"), 0644)
}

// clearSessionClaim removes a provisional reservation file.
func clearSessionClaim(cfg RegistryConfig, sessionID string) {
	_ = os.Remove(sessionClaimPath(cfg, sessionID))
}

func readSessionClaimPID(cfg RegistryConfig, sessionID string) (int, bool) {
	data, err := os.ReadFile(sessionClaimPath(cfg, sessionID))
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

// sessionIDInUse reports whether a session id is reserved or live.
func sessionIDInUse(cfg RegistryConfig, sessionID string) bool {
	if entry, err := ReadRegistry(cfg, sessionID); err == nil {
		if entry.ListenAddr != "" && tcpReachable(entry.ListenAddr) {
			return true
		}
		if entry.PID > 0 && processAlive(entry.PID) {
			return true
		}
	}
	if pid, ok := readSessionClaimPID(cfg, sessionID); ok && processAlive(pid) {
		return true
	}
	return false
}

// SessionIDInUse reports whether a session id is reserved or held by a live
// registry entry (reachable listen address and/or alive serve PID).
func SessionIDInUse(cfg RegistryConfig, sessionID string) bool {
	return sessionIDInUse(cfg, sessionID)
}

// pruneStaleSessionID drops unreachable registry entries and dead claims for id.
func pruneStaleSessionID(cfg RegistryConfig, sessionID string) error {
	if _, err := os.Stat(registryPath(cfg, sessionID)); err == nil {
		entry, readErr := ReadRegistry(cfg, sessionID)
		if readErr == nil {
			if entry.ListenAddr != "" && tcpReachable(entry.ListenAddr) {
				return nil // still live; caller checks inUse
			}
			if entry.PID > 0 && processAlive(entry.PID) {
				return nil
			}
			RemoveRegistry(cfg, sessionID)
		} else {
			// Corrupt or incomplete json — remove so the id can be reused.
			RemoveRegistry(cfg, sessionID)
		}
	}
	if pid, ok := readSessionClaimPID(cfg, sessionID); ok {
		if !processAlive(pid) {
			clearSessionClaim(cfg, sessionID)
		}
	}
	return nil
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

// WriteRegistry creates the registry file for a live session and drops any
// provisional claim for the same id.
func WriteRegistry(cfg RegistryConfig, entry RegistryEntry) error {
	if err := os.MkdirAll(RegistryDir(cfg), 0755); err != nil {
		return err
	}
	if entry.CreatedAt == "" {
		entry.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if err := os.WriteFile(registryPath(cfg, entry.SessionID), data, 0644); err != nil {
		return err
	}
	clearSessionClaim(cfg, entry.SessionID)
	return nil
}

// ReadRegistry loads a session registry entry.
func ReadRegistry(cfg RegistryConfig, sessionID string) (*RegistryEntry, error) {
	data, err := os.ReadFile(registryPath(cfg, sessionID))
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

// RemoveRegistry deletes the registry file and any claim for a session.
func RemoveRegistry(cfg RegistryConfig, sessionID string) {
	_ = os.Remove(registryPath(cfg, sessionID))
	clearSessionClaim(cfg, sessionID)
}

// RemoveRegistryIfMatch deletes the registry file only when the on-disk entry still
// belongs to the caller (same listen address and pid). Stale __serve__ cleanup can
// otherwise delete a newer session that reused the same session id.
func RemoveRegistryIfMatch(cfg RegistryConfig, sessionID, listenAddr string, pid int) {
	entry, err := ReadRegistry(cfg, sessionID)
	if err != nil {
		return
	}
	if entry.ListenAddr != listenAddr || entry.PID != pid {
		return
	}
	RemoveRegistry(cfg, sessionID)
}

// ListRegistryEntries returns all registry entries, optionally pruning unreachable ones.
func ListRegistryEntries(cfg RegistryConfig, prune bool) ([]RegistryEntry, error) {
	dir := RegistryDir(cfg)
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
		if prune && strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".claim") {
			id := strings.TrimSuffix(strings.TrimPrefix(name, "."), ".claim")
			if pid, ok := readSessionClaimPID(cfg, id); ok && !processAlive(pid) {
				clearSessionClaim(cfg, id)
			}
			continue
		}
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		entry, err := ReadRegistry(cfg, id)
		if err != nil {
			continue
		}
		if prune && !tcpReachable(entry.ListenAddr) {
			RemoveRegistryIfMatch(cfg, id, entry.ListenAddr, entry.PID)
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

// WaitForRegistryEntry polls until a session registry entry is available.
func WaitForRegistryEntry(cfg RegistryConfig, sessionID string, timeout time.Duration) (*RegistryEntry, error) {
	return waitForRegistryEntry(cfg, sessionID, timeout)
}

func waitForRegistryEntry(cfg RegistryConfig, sessionID string, timeout time.Duration) (*RegistryEntry, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		entry, err := ReadRegistry(cfg, sessionID)
		if err == nil && entry.ListenAddr != "" {
			return entry, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, fmt.Errorf("timeout waiting for registry entry %s", sessionID)
}

// LookupAcrossSubdirs searches subdirs under home for a session id.
// Returns the entry, the subdir where it was found, or an error.
func LookupAcrossSubdirs(home string, subdirs []string, sessionID string) (*RegistryEntry, string, error) {
	var lastErr error
	for _, subdir := range subdirs {
		cfg := RegistryConfig{Home: home, Subdir: subdir}
		entry, err := ReadRegistry(cfg, sessionID)
		if err == nil {
			return entry, subdir, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, "", lastErr
	}
	return nil, "", fmt.Errorf("tty-watch session %s not found", sessionID)
}

// TCPReachable reports whether a TCP listen address accepts connections.
func TCPReachable(addr string) bool {
	return tcpReachable(addr)
}

// ProcessAlive reports whether pid responds to signal 0.
func ProcessAlive(pid int) bool {
	return processAlive(pid)
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}