# Scenario

**Feature**: O3 — prompt-only discovery fallback when session cwd ≠ agent-run workspace

```
GROK_HOME seeded with session under encoded OTHER cwd
  (summary.info.cwd = /Users/other/path/for-prompt-fallback)
  agent-run workspace = TempDir (≠ session cwd)
  same open prompt text + created_at ≥ runStart
  -> agent-run run --open "prompt fallback cwd mismatch"
  -> workspace-keyed DiscoverSession misses
  -> scanAllSessionsForPrompt matches prompt + created_at
  -> bind sets runner_session_id; stderr grok session; exit 0
```

Locks the production fallback used when grok records a different cwd than
agent-run's `--dir` / workspace (e.g. `/tmp` vs `/private/tmp` style mismatch,
or a project picker that opens under another tree).

## Preconditions

- Preseed session under mismatched cwd; do **not** set session-id hook (pure prompt fallback).
- `GrokSessionCwd` harness field selects encoded path + summary cwd.

## Steps

1. Seed `GROK_HOME/sessions/<encoded-other-cwd>/<uuid>/{summary.json,updates.jsonl}` with matching prompt.
2. Leave agent-run workspace as TempDir (cmd.Dir); open with same prompt.
3. Assert bind success via prompt fallback.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"
)

const promptFallbackUUID = "550e8400-e29b-41d4-a716-446655440812"
const promptFallbackOtherCwd = "/Users/other/path/for-prompt-fallback"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prompt := "prompt fallback cwd mismatch"
	req.OpenPrompt = prompt
	req.InitialPrompt = prompt
	req.GrokHome = filepath.Join(req.TempDir, "grok-home-cwd-mismatch")
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return err
	}
	req.GrokSessionCwd = promptFallbackOtherCwd
	// Do not set GrokSessionUUID env hook — discovery must use prompt fallback.
	req.GrokSessionUUID = ""
	req.GrokUpdatesPath = writeFakeGrokSessionDirAtCwd(t, req.GrokHome, promptFallbackOtherCwd, promptFallbackUUID, prompt)
	// Sanity: agent workspace must differ from session cwd encoding.
	if encodedGrokCwd(req.Workspace) == encodedGrokCwd(promptFallbackOtherCwd) {
		return fmt.Errorf("test setup: workspace encoding collides with other cwd")
	}
	req.GrokTTYCommand = fakeTUIHoldSeconds(3)
	req.OpenInstantAttach = true
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open", prompt}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
