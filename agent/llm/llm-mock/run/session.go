package run

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FindNewestGrokSessionEvents returns the newest session dir and events.jsonl lines
// under grokHome/sessions/<url-encoded-abs-workDir>/.
func FindNewestGrokSessionEvents(grokHome, workDir string) (sessionDir, eventsPath string, lines []string, err error) {
	abs, err := workDirForSessionEncoding(workDir)
	if err != nil {
		return "", "", nil, err
	}
	encoded := url.PathEscape(abs)
	sessionsRoot := filepath.Join(grokHome, "sessions", encoded)
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		return "", "", nil, fmt.Errorf("read sessions root %s: %w", sessionsRoot, err)
	}

	type candidate struct {
		dir   string
		mtime int64
	}
	var candidates []candidate
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		dir := filepath.Join(sessionsRoot, ent.Name())
		eventsFile := filepath.Join(dir, "events.jsonl")
		if st, statErr := os.Stat(eventsFile); statErr == nil {
			candidates = append(candidates, candidate{dir: dir, mtime: st.ModTime().UnixNano()})
			continue
		}
		if st, statErr := os.Stat(dir); statErr == nil {
			candidates = append(candidates, candidate{dir: dir, mtime: st.ModTime().UnixNano()})
		}
	}
	if len(candidates) == 0 {
		return "", "", nil, fmt.Errorf("no session dirs under %s", sessionsRoot)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].mtime > candidates[j].mtime
	})
	sessionDir = candidates[0].dir
	eventsPath = filepath.Join(sessionDir, "events.jsonl")
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		return sessionDir, eventsPath, nil, fmt.Errorf("read events.jsonl: %w", err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return sessionDir, eventsPath, lines, scanner.Err()
}

// mirrorSessionsForWorkDir copies grok session trees from the canonical cwd path
// grok uses internally to the caller-provided workDir encoding expected by tests.
// On macOS /var is a symlink to /private/var; grok resolves symlinks when naming
// session dirs but doctest helpers PathEscape the unresolved orchestrator cwd.
// MirrorSessionsForWorkDirWithRetry copies session trees to the unresolved workDir
// encoding used by doctest helpers when grok stores sessions under a canonical path.
func MirrorSessionsForWorkDirWithRetry(grokHome, workDir string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if err := mirrorSessionsForWorkDir(grokHome, workDir); err != nil {
			return err
		}
		if mirroredSessionsReady(grokHome, workDir) {
			return nil
		}
		if time.Now().After(deadline) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func mirroredSessionsReady(grokHome, workDir string) bool {
	abs, err := workDirForSessionEncoding(workDir)
	if err != nil {
		return false
	}
	// This retry exists to make the caller-facing (unresolved) encoding ready.
	// A source-only event is not enough: returning then can race a late
	// link/copy and leave consumers with a missing events.jsonl at toEnc.
	ready, _ := sessionRootHasEventsForEncoding(grokHome, url.PathEscape(abs))
	return ready
}

func sessionRootHasEventsForEncoding(grokHome, encoded string) (bool, string) {
	target := filepath.Join(grokHome, "sessions", encoded)
	entries, err := os.ReadDir(target)
	if err != nil {
		return false, target
	}
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(target, ent.Name(), "events.jsonl")); err == nil {
			return true, target
		}
	}
	return false, target
}

func mirrorSessionsForWorkDir(grokHome, workDir string) error {
	abs, err := workDirForSessionEncoding(workDir)
	if err != nil {
		return err
	}
	toEnc := url.PathEscape(abs)
	toRoot := filepath.Join(grokHome, "sessions", toEnc)
	if _, err := os.Lstat(toRoot); err == nil {
		runGrokDebugf("mirrorSessionsForWorkDir: toRoot already exists %s", toRoot)
		return nil
	}

	fromRoot := filepath.Join(grokHome, "sessions", grokSessionEncoding(abs))
	if _, err := os.Stat(fromRoot); err == nil {
		runGrokDebugf("mirrorSessionsForWorkDir: link/copy %s -> %s", fromRoot, toRoot)
		return linkOrCopySessionRoot(fromRoot, toRoot)
	}

	// Fall back: grok may store sessions under any encoded cwd variant.
	sessionsRoot := filepath.Join(grokHome, "sessions")
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		return nil
	}
	for _, ent := range entries {
		if !ent.IsDir() || ent.Name() == toEnc {
			continue
		}
		candidate := filepath.Join(sessionsRoot, ent.Name())
		if !sessionRootHasEvents(candidate) {
			runGrokDebugf("mirrorSessionsForWorkDir: skip candidate %s (no events.jsonl)", candidate)
			continue
		}
		runGrokDebugf("mirrorSessionsForWorkDir: fallback link/copy %s -> %s", candidate, toRoot)
		return linkOrCopySessionRoot(candidate, toRoot)
	}
	runGrokDebugf("mirrorSessionsForWorkDir: no source found for toEnc=%s grokEnc=%s", toEnc, grokSessionEncoding(abs))
	return nil
}

func linkOrCopySessionRoot(fromRoot, toRoot string) error {
	if _, err := os.Lstat(toRoot); err == nil {
		return nil
	}
	if err := os.Symlink(fromRoot, toRoot); err == nil {
		return nil
	}
	return copyDir(fromRoot, toRoot)
}

func sessionRootHasEvents(sessionRoot string) bool {
	entries, err := os.ReadDir(sessionRoot)
	if err != nil {
		return false
	}
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(sessionRoot, ent.Name(), "events.jsonl")); err == nil {
			return true
		}
	}
	return false
}

// sessionsHaveAnyEvents reports whether any session under grokHome/sessions has events.jsonl.
func sessionsHaveAnyEvents(grokHome string) bool {
	sessionsRoot := filepath.Join(grokHome, "sessions")
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		return false
	}
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		if sessionRootHasEvents(filepath.Join(sessionsRoot, ent.Name())) {
			return true
		}
	}
	return false
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// workDirForSessionEncoding returns the cwd path doctest helpers encode with
// url.PathEscape. os.Getwd on macOS may return /private/var while callers pass
// /var; normalize so orchestrator and harness agree on session directory names.
func workDirForSessionEncoding(workDir string) (string, error) {
	abs, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("abs workdir: %w", err)
	}
	return denormalizePrivatePath(abs), nil
}

func denormalizePrivatePath(path string) string {
	if strings.HasPrefix(path, "/private/var/") {
		return "/var/" + strings.TrimPrefix(path, "/private/var/")
	}
	if strings.HasPrefix(path, "/private/tmp/") {
		return "/tmp/" + strings.TrimPrefix(path, "/private/tmp/")
	}
	return path
}

// grokSessionEncoding returns the cwd encoding grok uses for session directories.
// Grok resolves symlinks (e.g. macOS /var -> /private/var) before PathEscape.
func grokSessionEncoding(abs string) string {
	if canonical, err := filepath.EvalSymlinks(abs); err == nil && canonical != abs {
		return url.PathEscape(canonical)
	}
	enc := url.PathEscape(abs)
	if strings.HasPrefix(enc, "%2Fvar%2F") {
		return strings.Replace(enc, "%2Fvar%2F", "%2Fprivate%2Fvar%2F", 1)
	}
	if strings.HasPrefix(enc, "%2Ftmp%2F") {
		return strings.Replace(enc, "%2Ftmp%2F", "%2Fprivate%2Ftmp%2F", 1)
	}
	return enc
}
