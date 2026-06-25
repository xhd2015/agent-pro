# Scenario

**Feature**: subagent SessionLayout controls flat session directory artifacts

```
# caller provides flat Dir; subagent writes configured paths under it
test harness -> subagent.Run(SessionLayout.Dir) -> inspect session dir files
```

## Preconditions

- Module root is `agent-pro-task-hub` (four levels above `session-layout/`).
- `fake-codex` builds from `./cmd/fake-codex`.
- Implementation may be absent (tests expect RED).

## Steps

1. Resolve module root from `DOCTEST_ROOT`.
2. Build `fake-codex` binary into temp dir.
3. Descendant `Setup` creates flat session dir and configures `SessionLayout`.

## Context

- Default agent runner: `fake-codex`
- Session ID for flat-dir leaves: `gen_layout_flat_test`

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func moduleRoot(t *testing.T) string {
	t.Helper()
	return filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", "..", "..", ".."))
}

func Setup(t *testing.T, req *Request) error {
	req.TempDir = t.TempDir()
	req.HomeDir = filepath.Join(req.TempDir, "home")
	if err := os.MkdirAll(req.HomeDir, 0755); err != nil {
		return err
	}
	req.FakeCodexBin = filepath.Join(req.TempDir, "fake-codex")
	buildFakeCodex(t, moduleRoot(t), req.FakeCodexBin)
	if req.AgentRunner == "" {
		req.AgentRunner = "fake-codex"
	}
	if req.Prompt == "" {
		req.Prompt = "run layout test"
	}
	return nil
}
```
