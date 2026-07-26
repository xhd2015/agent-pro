# Scenario

**Feature**: `agent-run run --dir DIR` sets the session workspace independently of process cwd

```
# process cwd may differ from workspace
cwd=TempDir
agent-run run --dir <workspace> --agent-runner fake-codex "hi"
  -> sessions/fake-codex/<id>/meta.json workspace = abs(workspace)
  -> workspace is not process cwd when they differ

# omitted --dir → workspace defaults to process cwd (Getwd)
# invalid --dir → non-zero; clear missing / not-a-directory error
# run --help documents --dir
```

## Preconditions

- Inherits root binary build (`agent-run`, `fake-codex`) and `run` grouping args
  (`run --agent-runner fake-codex`).
- Each leaf keeps isolated `AGENT_RUN_HOME` under `req.TempDir`.
- Process cwd for `runAgentRun` is `req.TempDir` (root harness `cmd.Dir`).

## Steps

1. Grouping exposes helpers to read `meta.workspace` and compare paths canonically.
2. Leaves create fixture dirs/files under `req.TempDir`, append `--dir` / prompt args,
   and assert meta + exit.

## Context

- Flag form under test: space-separated `--dir PATH` (library also accepts `=` form).
- Relative paths resolve against process cwd then clean (EvalSymlinks preferred when possible).
- User-facing help stdout ends with trailing `\n`.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Inherit run/ prefix: run --agent-runner fake-codex
	// frontend-agent-run/dist is gitignored; ensure //go:embed can compile.
	if err := ensureStubDistForDir(filepath.Join(req.RepoRoot, "frontend-agent-run", "dist")); err != nil {
		return err
	}
	_ = ensureStubDistForDir(filepath.Join(req.RepoRoot, "frontend", "dist"))
	ensureDirHelpersUsed()
	return nil
}

// ensureStubDistForDir makes sure distDir has at least one embeddable file so
// //go:embed dist compiles when the real SPA dist is absent (gitignored).
func ensureStubDistForDir(distDir string) error {
	entries, statErr := os.ReadDir(distDir)
	if statErr == nil {
		for _, e := range entries {
			if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				return nil
			}
		}
	}
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(distDir, "index.html"), []byte("stub\n"), 0644)
}

func canonicalPath(path string) string {
	path = filepath.Clean(path)
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return path
}

func pathsEqual(a, b string) bool {
	return canonicalPath(a) == canonicalPath(b)
}

// findSessionMetaJSON walks AGENT_RUN_HOME/sessions for the first meta.json.
func findSessionMetaJSON(t *testing.T, home string) (path string, meta map[string]any) {
	t.Helper()
	sessionsRoot := filepath.Join(home, "sessions")
	var found string
	_ = filepath.Walk(sessionsRoot, func(p string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		if info.Name() == "meta.json" && found == "" {
			found = p
		}
		return nil
	})
	if found == "" {
		t.Fatalf("meta.json not found under %s", sessionsRoot)
	}
	data, err := os.ReadFile(found)
	if err != nil {
		t.Fatalf("read %s: %v", found, err)
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("parse %s: %v\n%s", found, err, string(data))
	}
	return found, meta
}

func sessionWorkspace(t *testing.T, home string) string {
	t.Helper()
	_, meta := findSessionMetaJSON(t, home)
	ws, _ := meta["workspace"].(string)
	return strings.TrimSpace(ws)
}

func assertSessionWorkspace(t *testing.T, home, want string) {
	t.Helper()
	got := sessionWorkspace(t, home)
	if got == "" {
		t.Fatalf("meta.workspace empty; want %q", want)
	}
	if !pathsEqual(got, want) {
		t.Fatalf("meta.workspace = %q, want %q (canonical %q vs %q)",
			got, want, canonicalPath(got), canonicalPath(want))
	}
}

func processCwd(req *Request) string {
	// Root runAgentRun uses cmd.Dir = req.TempDir.
	return req.TempDir
}

func ensureDirHelpersUsed() {
	_ = ensureStubDistForDir
	_ = canonicalPath
	_ = pathsEqual
	_ = findSessionMetaJSON
	_ = sessionWorkspace
	_ = assertSessionWorkspace
	_ = processCwd
}
```
