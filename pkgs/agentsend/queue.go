package agentsend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Entry is one pending send-queue message.
type Entry struct {
	ID                string    `json:"id"`
	Text              string    `json:"text"`
	TerminalSessionID string    `json:"terminal_session_id"`
	Runner            string    `json:"runner"`
	EnqueuedAt        time.Time `json:"enqueued_at"`
}

// Session identifies a live terminal send queue.
type Session struct {
	Home              string
	Runner            string
	TerminalSessionID string
	ListenAddr        string
}

func queueDir(home, runner string) string {
	return filepath.Join(home, "send-queue", runner)
}

func queuePath(home, runner, terminalSessionID string) string {
	return filepath.Join(queueDir(home, runner), terminalSessionID+".jsonl")
}

func lockPath(home, runner, terminalSessionID string) string {
	return filepath.Join(queueDir(home, runner), terminalSessionID+".lock")
}

func seqPath(home, runner, terminalSessionID string) string {
	return filepath.Join(queueDir(home, runner), terminalSessionID+".seq")
}

func ensureQueueDir(home, runner string) error {
	return os.MkdirAll(queueDir(home, runner), 0755)
}

func acquireLock(home, runner, terminalSessionID string) (func(), error) {
	if err := ensureQueueDir(home, runner); err != nil {
		return nil, err
	}
	path := lockPath(home, runner, terminalSessionID)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, err
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}

func readEntries(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var entries []Entry
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func writeEntries(path string, entries []Entry) error {
	if len(entries) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	var b strings.Builder
	for i, e := range entries {
		if i > 0 {
			b.WriteByte('\n')
		}
		line, err := json.Marshal(e)
		if err != nil {
			return err
		}
		b.Write(line)
	}
	return os.WriteFile(path, []byte(b.String()+"\n"), 0644)
}

func containsID(entries []Entry, id string) bool {
	for _, e := range entries {
		if e.ID == id {
			return true
		}
	}
	return false
}

func readSeq(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	var n int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &n); err != nil {
		return 0, err
	}
	return n, nil
}

func writeSeq(path string, n int) error {
	return os.WriteFile(path, []byte(fmt.Sprintf("%d\n", n)), 0644)
}

func nextMessageID(home, runner, terminalSessionID string, entries []Entry) (string, error) {
	maxN, err := readSeq(seqPath(home, runner, terminalSessionID))
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		var n int
		if _, err := fmt.Sscanf(e.ID, "msg_%d", &n); err == nil && n > maxN {
			maxN = n
		}
	}
	next := maxN + 1
	if err := writeSeq(seqPath(home, runner, terminalSessionID), next); err != nil {
		return "", err
	}
	return fmt.Sprintf("msg_%d", next), nil
}

func appendEntryLocked(path string, entry Entry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}
	return nil
}

func removeEntryByID(path string, id string) (bool, error) {
	entries, err := readEntries(path)
	if err != nil {
		return false, err
	}
	found := false
	kept := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if e.ID == id {
			found = true
			continue
		}
		kept = append(kept, e)
	}
	if !found {
		return false, nil
	}
	if err := writeEntries(path, kept); err != nil {
		return false, err
	}
	return true, nil
}

func queueContains(home, runner, terminalSessionID, msgID string) (bool, error) {
	entries, err := readEntries(queuePath(home, runner, terminalSessionID))
	if err != nil {
		return false, err
	}
	return containsID(entries, msgID), nil
}

func loadHead(path string) (*Entry, error) {
	entries, err := readEntries(path)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	head := entries[0]
	return &head, nil
}

func dequeueHead(path string) error {
	entries, err := readEntries(path)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	return writeEntries(path, entries[1:])
}

