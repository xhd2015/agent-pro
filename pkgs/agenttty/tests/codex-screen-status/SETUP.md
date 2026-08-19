# Scenario

**Feature**: `codex-tty` `DetectScreenStatus` on finished live Codex chrome

```
tty snapshot / crime-scene fixture
  -> agenttty.Get("codex-tty").DetectScreenStatus
  -> idle
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agenttty` registers `codex-tty`
  with `DetectScreenStatus` and `CheckWritable`.
- Fixtures live under this tree’s `testdata/` (rendered snapshot text).
- `d.DOCTEST_ROOT` is this tree; cwd is undetermined — join paths from `d`.
- Parallel-safe: no `os.Setenv` / `t.Setenv` / `Chdir`.
- Do not change sealed `codex-writable` leaves that still allow `banner|idle`.

## Steps

1. Root `Setup` sets `TestdataDir`.
2. Grouping `finished-prompt/` records the chrome class.
3. Leaf sets `Fixture`.
4. `Run` calls `DetectScreenStatus` (+ occupancy + writable).
5. Leaf `Assert` requires `screen=idle`.

## Context

- Crime scene: `~/.sandbox/transcripts/2026-08-18T105032Z-crime-scene-codex-tty-idle-exit.md`.
- Desired: live `›` without `CODEX_TTY_BANNER` is `idle` (same as grok
  `detectGrokScreenStatus` after the boxed-chrome alignment).
- `SampleIsIdle` needs `screen==idle`; writable/sendable is already idle.

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const (
	fixtureHostV10          = "idle-probe-v10-host.txt"
	fixtureMockModelDefault = "idle-probe-scene-mock-model.txt"
	fixtureEmptyGlued       = "codex-0.147-empty-glued.txt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.TestdataDir == "" {
		req.TestdataDir = filepath.Join(d.DOCTEST_ROOT, "testdata")
	}
	return nil
}

func assertScreenIdle(t *testing.T, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if !strings.Contains(resp.Text, "›") && !strings.Contains(resp.Text, "\u203a") {
		t.Fatal("fixture must contain Codex ›")
	}
	if strings.Contains(resp.Text, "CODEX_TTY_BANNER") {
		t.Fatal("live chrome fixture must not contain CODEX_TTY_BANNER")
	}
	if resp.Screen != "idle" {
		t.Fatalf("DetectScreenStatus=%q want idle (live › chrome)", resp.Screen)
	}
}
```
