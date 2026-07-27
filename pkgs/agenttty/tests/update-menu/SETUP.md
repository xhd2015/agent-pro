# Scenario

**Feature**: classify signed update-modal snapshots (menu vs residual banner)

```
# classify path
signed snapshot fixture
  -> IsBlockingUpdateMenu / UpdateMenuSelection / codex-tty CheckWritable
```

## Preconditions

1. Fixtures are **copied** into `testdata/update-modal-skip/` (self-contained; no
   ai-critic path dependency).
2. Production helpers on `github.com/xhd2015/agent-pro/pkgs/agenttty`:
   - `IsBlockingUpdateMenu(text string) bool`
   - `UpdateMenuSelection(text string) string`
3. `checkCodexWritable` narrows update gate to **blocking menu**, not bare banner.

## Steps

1. Root `Setup` resolves default fixtures dir.
2. Leaf sets `FixtureFile` and optional `StripModelLoading`.
3. Root `Run` classifies and returns writable status.

## Context

Source-module ownership for classifiers. End-to-end FetchStatus Skip protocol
lives under `agent/codex/tty/tests/update-modal-skip/`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.RepoRoot == "" {
		req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", "..", "..", ".."))
	}
	if req.FixturesDir == "" {
		req.FixturesDir = filepath.Join(d.DOCTEST_ROOT, "testdata", "update-modal-skip")
	}
	return nil
}
```
