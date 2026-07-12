# Scenario

**Feature**: nested→flat agent-run session migration CLI

```
go build cmd/migrate-sessions -> migrate-sessions --home H [--dry-run] [--backup-dir D]
nested sessions/<runner>/<id>/ -> flat sessions/<id>/ + .layout v2 + backup
```

## Preconditions

- Source package path is `./cmd/migrate-sessions` (implementer creates it).
- Each leaf uses an isolated temp home with fixture nested or flat trees.
- Registry / send-queue trees must not be modified by migration.

## Steps

1. Root `Setup` builds the migrator binary into the temp bin dir.
2. Grouping/leaf `Setup` seeds nested or flat session fixtures under `req.Home`.
3. `Run` executes the binary with `--home` and optional flags.
4. Leaf `Assert` checks exit code, filesystem layout, backup, and `.layout`.

## Context

- Fixtures write raw `meta.json` / `events.jsonl` (do not depend on flat Store API).
- Collision timestamps use RFC3339 `updated_at` fields.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	if err := os.MkdirAll(req.Home, 0o755); err != nil {
		return fmt.Errorf("mkdir home: %w", err)
	}
	req.Bin = filepath.Join(req.TempDir, "bin", "migrate-sessions")
	if err := os.MkdirAll(filepath.Dir(req.Bin), 0o755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	build := exec.Command("go", "build", "-o", req.Bin, "./cmd/migrate-sessions")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build migrate-sessions: %w\n%s", err, string(out))
	}
	req.Env = append(req.Env, "AGENT_RUN_HOME="+req.Home)
	return nil
}

func writeNestedSession(t *testing.T, home, runner, sessionID, status, updatedAt, eventText string) {
	t.Helper()
	dir := filepath.Join(home, "sessions", runner, sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir nested session: %v", err)
	}
	meta := map[string]any{
		"runner":     runner,
		"session_id": sessionID,
		"status":     status,
		"created_at": updatedAt,
		"updated_at": updatedAt,
	}
	// deliberately omit runner sometimes for "ensure meta.runner set from path" cases
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), data, 0o644); err != nil {
		t.Fatalf("write meta: %v", err)
	}
	if eventText != "" {
		line := fmt.Sprintf(`{"type":"message","text":%q}`+"\n", eventText)
		if err := os.WriteFile(filepath.Join(dir, "events.jsonl"), []byte(line), 0o644); err != nil {
			t.Fatalf("write events: %v", err)
		}
	}
}

func writeNestedSessionNoRunnerField(t *testing.T, home, runner, sessionID, status, updatedAt string) {
	t.Helper()
	dir := filepath.Join(home, "sessions", runner, sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	meta := map[string]any{
		"session_id": sessionID,
		"status":     status,
		"created_at": updatedAt,
		"updated_at": updatedAt,
	}
	data, _ := json.Marshal(meta)
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), data, 0o644); err != nil {
		t.Fatalf("write meta: %v", err)
	}
}

func writeFlatSession(t *testing.T, home, sessionID, runner, status, updatedAt string) {
	t.Helper()
	dir := filepath.Join(home, "sessions", sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir flat: %v", err)
	}
	meta := map[string]any{
		"runner":     runner,
		"session_id": sessionID,
		"status":     status,
		"created_at": updatedAt,
		"updated_at": updatedAt,
	}
	data, _ := json.Marshal(meta)
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), data, 0o644); err != nil {
		t.Fatalf("write meta: %v", err)
	}
}

func writeLayoutV2(t *testing.T, home string) {
	t.Helper()
	dir := filepath.Join(home, "sessions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".layout"), []byte(`{"version":2}`+"\n"), 0o644); err != nil {
		t.Fatalf("write .layout: %v", err)
	}
}

func readMetaRunner(t *testing.T, metaPath string) string {
	t.Helper()
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta %s: %v", metaPath, err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal meta: %v", err)
	}
	s, _ := m["runner"].(string)
	return s
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s\nstdout:\n%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	st, err := os.Stat(path)
	if err != nil || !st.IsDir() {
		t.Fatalf("expected directory %q: %v", path, err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %q: %v", path, err)
	}
}

func assertPathMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected path missing: %q", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %q: %v", path, err)
	}
}

func layoutVersion(t *testing.T, home string) int {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, "sessions", ".layout"))
	if err != nil {
		t.Fatalf("read .layout: %v", err)
	}
	var m struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse .layout: %v (%s)", err, string(data))
	}
	return m.Version
}

func backupDirsUnder(t *testing.T, home string) []string {
	t.Helper()
	// default backup: $HOME_DIR/backups/sessions-<timestamp>/
	root := filepath.Join(home, "backups")
	ents, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read backups: %v", err)
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() && strings.HasPrefix(e.Name(), "sessions-") {
			out = append(out, filepath.Join(root, e.Name()))
		}
	}
	return out
}
```
