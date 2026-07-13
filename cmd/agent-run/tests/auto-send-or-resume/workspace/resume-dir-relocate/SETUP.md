# Scenario

**Feature**: on resume (and auto→resume), when `--dir` is set and runner is
`grok-tty`, compare resolved `--dir` to Grok session `summary.json` `info.cwd`
for `meta.runner_session_id` under effective Grok home
(`--agent-runner-config-home`).

```
--dir unset                         -> unchanged (existing meta.workspace rules)
--dir == Grok info.cwd (canonical)  -> OK; no RelocateCWD
--dir ≠ Grok cwd, no allow flag     -> exit 1; stderr mismatch + allow flag hint;
                                       Grok session dir NOT moved
--dir ≠ Grok cwd + --allow-relocate-resume-session-dir
                                    -> stderr warning; sessions.RelocateCWD;
                                       meta.workspace = --dir; resume continues
MODE=run / non-grok-tty             -> no Grok cwd check
```

Flag (exact): `--allow-relocate-resume-session-dir`  
Wired on: `run` (auto-send path) and `resume` subcommand.  
Path equality: Abs + EvalSymlinks + `/private` prefix variants equal.

## Preconditions

- Inherits suite + `workspace/` Setup (created-ws, cli-cwd; flat AGENT_RUN_HOME).
- Seed inactive Grok session (no `active_sessions.json` entry for id) under:
  `$GROK_HOME/sessions/<url.PathEscape(abs_cwd)>/<runner_session_id>/summary.json`
  with `info.id` = runner_session_id and `info.cwd` = abs cwd.
- Pass `--agent-runner-config-home` pointing at fixture Grok home.
- Use argv/cwd recording runner (`installArgvCwdRunner`); no
  `AGENT_RUN_GROK_TTY_COMMAND`.
- `seedBoundExitedDeadTerminal` for MODE=resume / auto→resume.

## Steps

1. Grouping `Setup` allocates isolated Grok home under TempDir, clears ambient
   `AGENT_RUN_GROK_TTY_COMMAND`, and sets a default exec timeout.
2. Leaf creates old/new workspace dirs under TempDir as needed.
3. Leaf seeds Grok home + agent-run meta (bound+exited, runner_session_id).
4. Leaf installs fake runner and builds `resume` or auto Args with `--dir`
   and optional `--allow-relocate-resume-session-dir`.
5. Assert checks exit, stderr, Grok session path layout, summary `info.cwd`,
   and meta.workspace.

## Context

Helpers below encode cwd the same way as Grok / `sessions.RelocateCWD`
(`url.PathEscape(filepath.Abs(cwd))`). Encoding uses Abs **without**
EvalSymlinks so session keys match RelocateCWD.

```go
import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// absPathNoEval matches sessions.RelocateCWD Abs encoding (no EvalSymlinks).
func absPathNoEval(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("Abs %q: %v", p, err)
	}
	return abs
}

func encodeGrokCWDKey(t *testing.T, cwd string) string {
	t.Helper()
	return url.PathEscape(absPathNoEval(t, cwd))
}

func grokSessionDirAt(t *testing.T, grokHome, cwd, sessionID string) string {
	t.Helper()
	return filepath.Join(grokHome, "sessions", encodeGrokCWDKey(t, cwd), sessionID)
}

func ensureGrokHome(t *testing.T, req *Request) {
	t.Helper()
	if req.GrokHome == "" {
		req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	}
	if err := os.MkdirAll(filepath.Join(req.GrokHome, "sessions"), 0755); err != nil {
		t.Fatalf("mkdir grok sessions: %v", err)
	}
}

func writeJSONFile(t *testing.T, path string, v any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir parent %s: %v", path, err)
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// seedInactiveGrokSession writes summary.json under sessions/<encode(cwd)>/<id>/
// and ensures the session is inactive (empty active_sessions.json).
// Returns the absolute session directory path.
func seedInactiveGrokSession(t *testing.T, grokHome, cwd, sessionID string) string {
	t.Helper()
	absCwd := absPathNoEval(t, cwd)
	dir := grokSessionDirAt(t, grokHome, absCwd, sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir grok session dir: %v", err)
	}
	summary := map[string]any{
		"info": map[string]any{
			"id":  sessionID,
			"cwd": absCwd,
		},
		"generated_title": "resume-dir-relocate fixture",
		"created_at":      "2026-07-01T10:00:00.000Z",
		"updated_at":      "2026-07-01T11:00:00.000Z",
		"last_active_at":  "2026-07-01T11:00:00.000Z",
		"num_messages":    1,
		"num_chat_messages": 1,
	}
	writeJSONFile(t, filepath.Join(dir, "summary.json"), summary)
	// Marker so tests can confirm the directory moved (not only recreated).
	if err := os.WriteFile(filepath.Join(dir, "fixture-marker.txt"), []byte("relocate-v1\n"), 0644); err != nil {
		t.Fatalf("write fixture marker: %v", err)
	}
	// Explicit inactive list — RelocateCWD rejects active sessions.
	writeJSONFile(t, filepath.Join(grokHome, "active_sessions.json"), map[string]any{
		"sessions": []any{},
	})
	return dir
}

func summaryInfoCWD(t *testing.T, summaryPath string) string {
	t.Helper()
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read summary %s: %v", summaryPath, err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse summary: %v", err)
	}
	info, _ := m["info"].(map[string]any)
	if info == nil {
		t.Fatalf("summary missing info: %s", summaryPath)
	}
	cwd, _ := info["cwd"].(string)
	return cwd
}

func mustMkdirWS(t *testing.T, dir, marker string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(marker+"\n"), 0644); err != nil {
		t.Fatalf("write README in %s: %v", dir, err)
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func Setup(t *testing.T, req *Request) error {
	// Shared fixture for all relocate leaves: isolated Grok home, no ambient
	// TTY command hook (argv probes must see real --resume), default timeout.
	// Leaves still seed sessions / meta / --dir Args.
	ensureGrokHome(t, req)
	req.GrokTTYCommand = ""
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	if req.ExecTimeout <= 0 {
		req.ExecTimeout = 60 * time.Second
	}
	return nil
}
```
