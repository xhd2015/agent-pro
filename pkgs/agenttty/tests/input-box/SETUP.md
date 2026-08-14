# Scenario

**Feature**: classify live composer occupancy from snapshot / scrollback text

```
# last composer glyph on snapshot text
req.Scrollback | testdata fixture
  -> agenttty.DetectInputBox
  -> empty | occupied | unknown

# status printers share the library builder
DetectInputBox token -> InputBoxReport -> "input box: …" / input_box
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agenttty` will export a pure probe
  `DetectInputBox` on snapshot text and `InputBoxReport` for CLI tokens.
- Neither API exists yet (classic TDD; leaves compile/link RED).
- Locked live last-line shapes live under this tree’s `testdata/`.
- `d.DOCTEST_ROOT` is this tree; cwd is undetermined — join paths from `d`.
- No `os.Setenv` / `t.Setenv` / `Chdir` / hijack of `os.Stdout|Stderr|Stdin`.

## Steps

1. Root `Setup` sets `TestdataDir` and default `ProviderID=codex-tty`.
2. Grouping `Setup` records the composer family.
3. Leaf `Setup` injects `Scrollback` or a fixture basename.
4. `Run` calls `DetectInputBox` (+ `InputBoxReport`, + existing `CheckWritable`).
5. Leaf `Assert` checks `resp.InputBox` (and report / writable where relevant).

## Context

- `sendable` is yes / `sendable_state=idle` on **both** empty and occupied live
  Codex 0.147 snapshots. Occupancy is a new field, not an overload of sendable
  or `screen_status`.
- Footer marker is the exact substring ` medium · ` (spaces + middle dot U+00B7).
- Grok has no experiment lock; keep Grok conservative (padding-only → empty).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const (
	fixtureCodexEmptyGlued       = "codex-0.147-empty-glued.txt"
	fixtureCodexOccupiedSingle   = "codex-0.147-occupied-single.txt"
	fixtureCodexOccupiedMultiline = "codex-0.147-occupied-multiline.txt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.TestdataDir == "" {
		req.TestdataDir = filepath.Join(d.DOCTEST_ROOT, "testdata")
	}
	if req.ProviderID == "" {
		req.ProviderID = "codex-tty"
	}
	return nil
}

func assertInputBox(t *testing.T, resp *Response, err error, want string) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.InputBox != want {
		t.Fatalf("InputBox=%q want %q (human=%q json=%q)", resp.InputBox, want, resp.HumanLine, resp.JSONValue)
	}
}

func assertWritableIdle(t *testing.T, resp *Response, label string) {
	t.Helper()
	if !resp.WritableReady {
		t.Fatalf("%s: CheckWritable ready=false state=%q (occupancy must not flip sendable)", label, resp.WritableState)
	}
	if resp.WritableState != "idle" {
		t.Fatalf("%s: CheckWritable state=%q want idle", label, resp.WritableState)
	}
}
```
