# Scenario

**Feature**: file-backed agent run storage under `AGENT_RUN_HOME`

```
t.TempDir/.agent-run -> AGENT_RUN_HOME -> NewFileStore -> Store CRUD
config.json + sessions/<runner>/<id>/{meta.json,events.jsonl,messages.jsonl}
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentstorage` implements `Store` and `NewFileStore`.
- Each test uses an isolated temp directory; `AGENT_RUN_HOME` points at `filepath.Join(temp, ".agent-run")`.
- Tests call store methods directly (no CLI subprocess).

## Steps

1. Root `Setup` creates `req.TempDir` and `req.Home` under `t.TempDir()`.
2. Leaf `Setup` sets `req.Operation`, `req.Action`, runner/session fields.
3. `Run` opens the store, executes the operation, returns `Response`.
4. Leaf `Assert` checks outcomes.

## Context

- `OpenStore(t, req)` resolves home from `AGENT_RUN_HOME` (when set) and returns `(Store, homePath, error)`.
- `WriteEvent` appends a single `types.AgentEvent` via the store.
- `AssertHomeOnly` verifies every path in `FilesWritten` is under `home` prefix.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

var savedEnv = map[string]*string{}

func Setup(t *testing.T, req *Request) error {
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.Env = append(req.Env, "AGENT_RUN_HOME="+req.Home)
	if req.Runner == "" {
		req.Runner = "fake-opencode"
	}
	return nil
}

func openStore(t *testing.T, req *Request) (agentstorage.Store, string, error) {
	t.Helper()
	home := req.Home
	if v := os.Getenv("AGENT_RUN_HOME"); v != "" {
		home = v
	}
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		return nil, home, fmt.Errorf("NewFileStore(%q): %w", home, err)
	}
	return store, store.Home(), nil
}

func WriteEvent(t *testing.T, store agentstorage.Store, runner, sessionID, text string) {
	t.Helper()
	ev := types.AgentEvent{Type: types.ActionMessage, Text: text}
	if err := store.AppendEvent(runner, sessionID, ev); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}
}

func AssertHomeOnly(t *testing.T, home string, paths []string) {
	t.Helper()
	home = filepath.Clean(home)
	for _, p := range paths {
		p = filepath.Clean(p)
		if p == home || strings.HasPrefix(p, home+string(filepath.Separator)) {
			continue
		}
		t.Fatalf("path %q is outside home %q", p, home)
	}
}

func applyEnv(env []string) {
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]
		if old, ok := os.LookupEnv(key); ok {
			v := old
			savedEnv[key] = &v
		} else {
			savedEnv[key] = nil
		}
		os.Setenv(key, val)
	}
}

func restoreEnv(env []string) {
	seen := map[string]bool{}
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		if seen[key] {
			continue
		}
		seen[key] = true
		if old, ok := savedEnv[key]; ok {
			if old == nil {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, *old)
			}
		}
	}
	savedEnv = map[string]*string{}
}

type fileTracker struct {
	home  string
	paths []string
}

func newFileTracker(home string) *fileTracker {
	return &fileTracker{home: filepath.Clean(home)}
}

func scanFileTracker(ft *fileTracker) {
	_ = filepath.Walk(ft.home, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			ft.paths = append(ft.paths, filepath.Clean(path))
		}
		return nil
	})
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func assertEqual(t *testing.T, field string, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %v, want %v", field, got, want)
	}
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in %q", want, got)
	}
}
```