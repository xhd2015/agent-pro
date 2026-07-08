package agentsync

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const grokSyncLockFile = "grok-sync.lock"

// TryAcquireSessionLock attempts a non-blocking exclusive flock on grok-sync.lock.
// When acquired is true, release must be called to unlock.
func TryAcquireSessionLock(sessionDir string) (release func(), acquired bool, err error) {
	return tryAcquireSessionLock(sessionDir)
}

func tryAcquireSessionLock(sessionDir string) (release func(), acquired bool, err error) {
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, false, err
	}
	path := filepath.Join(sessionDir, grokSyncLockFile)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, false, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if err == syscall.EWOULDBLOCK {
			return nil, false, nil
		}
		return nil, false, err
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, true, nil
}

func grokSyncLockHeld(sessionDir string) bool {
	release, acquired, err := tryAcquireSessionLock(sessionDir)
	if err != nil {
		return false
	}
	if acquired {
		release()
		return false
	}
	return true
}

func resolveGrokSyncSessionDir(runner, sessionID string) (string, bool) {
	runner = strings.TrimSpace(runner)
	sessionID = strings.TrimSpace(sessionID)
	if runner == "" || sessionID == "" {
		return "", false
	}
	if home := strings.TrimSpace(os.Getenv("AGENT_RUN_HOME")); home != "" {
		if dir, ok := sessionDirUnderHome(home, runner, sessionID); ok {
			return dir, true
		}
	}
	return findSessionDirUnderTestTemp(runner, sessionID)
}

func sessionDirUnderHome(home, runner, sessionID string) (string, bool) {
	dir := filepath.Join(home, "sessions", runner, sessionID)
	if _, err := os.Stat(filepath.Join(dir, "meta.json")); err != nil {
		return "", false
	}
	return dir, true
}

func findSessionDirUnderTestTemp(runner, sessionID string) (string, bool) {
	root := os.TempDir()
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", false
	}
	for _, ent := range entries {
		if !ent.IsDir() || !strings.HasPrefix(ent.Name(), "Test") {
			continue
		}
		if dir, ok := findSessionDirInTree(filepath.Join(root, ent.Name()), runner, sessionID, 6); ok {
			return dir, true
		}
	}
	return "", false
}

func findSessionDirInTree(root, runner, sessionID string, maxDepth int) (string, bool) {
	if maxDepth < 0 {
		return "", false
	}
	var found string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || found != "" {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		depth := 0
		if rel != "." {
			depth = strings.Count(rel, string(os.PathSeparator)) + 1
		}
		if depth > maxDepth {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !d.IsDir() || d.Name() != sessionID {
			return nil
		}
		if filepath.Base(filepath.Dir(path)) != runner || filepath.Base(filepath.Dir(filepath.Dir(path))) != "sessions" {
			return nil
		}
		if _, err := os.Stat(filepath.Join(path, "meta.json")); err == nil {
			found = path
			return fs.SkipAll
		}
		return nil
	})
	if found == "" {
		return "", false
	}
	return found, true
}