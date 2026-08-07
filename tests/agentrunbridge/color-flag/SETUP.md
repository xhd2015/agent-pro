# Scenario

**Feature**: `RunOpts.Color` controls `--color` in `BuildArgs` argv

```
RunOpts{Color:true, Open, …} -> BuildArgs -> []string includes --color
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentrunbridge` exports
  `BuildArgs` and `RunOpts` with **`Color bool`** (RED until implementer adds
  the field + emission).
- Nested root: `d.DOCTEST_ROOT` is `tests/agentrunbridge/color-flag`.
- No real agent-run binary.

## Steps

1. Leaf fills open-profile opts with Color true.
2. `Run` returns `BuildArgs` result; assert exact argv.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.SessionID == "" {
		req.SessionID = "sess-color-1"
	}
	if req.Prompt == "" {
		req.Prompt = "with color"
	}
	if req.AgentRunner == "" {
		req.AgentRunner = "grok-tty"
	}
	return nil
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
}

func assertArgsEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("args len=%d want %d\ngot  %#v\nwant %#v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("args[%d]=%q want %q\ngot  %#v\nwant %#v", i, got[i], want[i], got, want)
		}
	}
}

func assertNoAPIError(t *testing.T, resp *Response) {
	t.Helper()
	if resp != nil && resp.ErrString != "" {
		t.Fatalf("unexpected API error: %s", resp.ErrString)
	}
}
```
