# Scenario

**Feature**: shared agent-run bridge library (`Run` + `RunInteractiveOpen`)

```
# pure argv
RunOpts -> BuildArgs -> []string (CLI args after binary)

# pure status
tty status stdout -> ParseTTYStatus / IsSessionReady -> ready?

# exec with fakes
LookPath + RunCommand + RunOutput hooks
  -> Run(RunOpts) | RunInteractiveOpen(InteractiveOpenOpts)
  -> optional wait-ready poll loop
  -> RunResult | error
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentrunbridge` exports:
  - `Run(opts RunOpts) (RunResult, error)`
  - `RunInteractiveOpen(opts InteractiveOpenOpts) (RunResult, error)`
  - `BuildArgs(opts RunOpts) []string`
  - `ParseTTYStatus(stdout string) (screen, sendable string)`
  - `IsSessionReady(stdout string) bool`
- Types: `RunOpts`, `RunResult`, `InteractiveOpenOpts` with fields per REQUIREMENT-DESIGN.
- Test hooks on opts: `LookPath`, `RunCommand`, `RunOutput` (and optional
  `ReadyTimeout` / `ReadyPollInterval` on interactive opts for short timeouts).
- **No real `agent-run` binary, PATH, or iTerm** — all unit leaves use fakes.
- Package does not exist yet (implementer creates it); leaves are RED until then.
- Argv uses equals form: `--session-id=`, `--agent-runner=`, `--dir=`.
- `BuildArgs` returns args **without** the binary name (first element is `run`).

## Steps

1. Root `Setup` seeds default session id and prompt used by most leaves.
2. Grouping `Setup` sets `req.Mode`.
3. Leaf `Setup` fills option fields and scripted hook fixtures.
4. `Run` calls package APIs; leaf `Assert` checks argv, status, or errors.

## Context

- Default fixtures: `SessionID=sess-bridge-1`, `Prompt=hello bridge`.
- Status fixtures: `statusReadyFixture()` / `statusNotReadyFixture()` in DOCTEST.md.
- Harness stores API errors on `Response.ErrString` (harness `error` is nil).

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.SessionID == "" {
		req.SessionID = "sess-bridge-1"
	}
	if req.Prompt == "" && req.Mode != "run" {
		// Leaves that need empty prompt set Prompt explicitly after root Setup.
		req.Prompt = "hello bridge"
	}
	return nil
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
}

func assertAPIError(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil || resp.ErrString == "" {
		t.Fatal("expected API error, got nil/empty")
	}
}

func assertNoAPIError(t *testing.T, resp *Response) {
	t.Helper()
	if resp != nil && resp.ErrString != "" {
		t.Fatalf("unexpected API error: %s", resp.ErrString)
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

func assertArgsEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("argv len: got %d %q, want %d %q", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("argv[%d]: got %q, want %q (full got=%q want=%q)", i, got[i], want[i], got, want)
		}
	}
}

func assertHasArg(t *testing.T, args []string, want string) {
	t.Helper()
	for _, a := range args {
		if a == want {
			return
		}
	}
	t.Fatalf("missing arg %q in %q", want, args)
}

func assertNotHasArg(t *testing.T, args []string, want string) {
	t.Helper()
	for _, a := range args {
		if a == want {
			t.Fatalf("unexpected arg %q in %q", want, args)
		}
	}
}

func assertNotHasArgPrefix(t *testing.T, args []string, prefix string) {
	t.Helper()
	for _, a := range args {
		if strings.HasPrefix(a, prefix) {
			t.Fatalf("unexpected arg with prefix %q: %q in %q", prefix, a, args)
		}
	}
}

func assertArgIndex(t *testing.T, args []string, want string) int {
	t.Helper()
	for i, a := range args {
		if a == want {
			return i
		}
	}
	t.Fatalf("missing arg %q in %q", want, args)
	return -1
}
```
