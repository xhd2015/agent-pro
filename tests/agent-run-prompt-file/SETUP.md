# Scenario

**Feature**: `agent-run run --prompt-file` resolves initial prompt (pure L2 + CLI wire)

```
# pure resolve (L2)
ResolveRunPrompt(positional, promptFile)
  -> prompt string | exclusive / missing error

# CLI surface (L2; pure file read — no process stdio swap)
read pkgs/agentruncli/run_cmd.go -> runHelp documents --prompt-file
scan pkgs/agentruncli -> flag literal registered
```

## Preconditions

- Package roots under agent-pro module:
  - `github.com/xhd2015/agent-pro/pkgs/agentruncli`
- Classic TDD: `ResolveRunPrompt` and `--prompt-file` may be missing
  (compile RED and/or assert RED) until implementer adds them.
- Network-free; no real agent-run binary spawn, iTerm, codex TTY, or PATH LookPath.
- Parallel-safe: no `t.Setenv` / `os.Setenv` / `t.Chdir` / process-global env
  mutation; **no process stdio reassignment**. File fixtures only under
  `d.DOCTEST_CASE` via `writeCaseFile` / `missingCasePath`.
- CLI help is a pure read of `pkgs/agentruncli/run_cmd.go` (holds `runHelp`);
  no `Handle` / process stdio capture.
- `d.DOCTEST_ROOT` is `tests/agent-run-prompt-file`; module root is found by
  walking for `module github.com/xhd2015/agent-pro`.
- Locked flag name: `--prompt-file` (not `--prompt-from-file`).

## Steps

1. Root `Setup` validates Request; trims Mode / paths lightly.
2. Grouping `Setup` sets `req.Mode`.
3. Leaf writes case fixtures under `d.DOCTEST_CASE` and fills Positional /
   PromptFile / SourceWireTarget.
4. `Run` resolves prompt, reads help source, or scans sources.
5. Leaf `Assert` checks outcomes.

## Context

- Fixture prompt body (after trim): `hello` (`fixturePromptBody`).
- Fixture file raw body: `"  hello\n"` (`fixturePromptFileRaw`) — documents
  TrimSpace policy matching positional.
- Fixture exclusive positional: `x` (`fixturePositional`).
- Flag token: `--prompt-file` (`flagPromptFile`).

```go
import (
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	if req == nil {
		return fmt.Errorf("nil Request")
	}
	_ = d
	req.Mode = strings.TrimSpace(req.Mode)
	req.SourceWireTarget = strings.TrimSpace(req.SourceWireTarget)
	// Positional / PromptFile: leaves set; do not invent defaults that hide exclusive cases.
	switch req.Mode {
	case "resolve":
		// Leaf owns Positional + PromptFile (and case files).
	case "cli_help":
		// Pure source read of run_cmd.go in Run; no Args needed.
	case "source_wire":
		// SourceWireTarget set by grouping/leaf.
	case "":
		// Leaf/grouping will set Mode.
	default:
		// Unknown modes fail in Run.
	}
	return nil
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
}

func assertNoAPIError(t *testing.T, resp *Response) {
	t.Helper()
	if resp != nil && resp.ErrString != "" {
		t.Fatalf("unexpected API error: %s", resp.ErrString)
	}
}

func assertAPIError(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil || resp.ErrString == "" {
		t.Fatal("expected ResolveRunPrompt error, got nil/empty")
	}
}

func assertEqual(t *testing.T, field string, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %#v, want %#v", field, got, want)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in %q", want, got)
	}
}

func assertNotContains(t *testing.T, got, forbidden string) {
	t.Helper()
	if strings.Contains(got, forbidden) {
		t.Fatalf("unexpected %q in %q", forbidden, got)
	}
}
```
