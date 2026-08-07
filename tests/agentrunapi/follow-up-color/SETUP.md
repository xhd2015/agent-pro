# Scenario

**Feature**: `FollowUpOpts.Color` controls `--color` on follow-up argv

```
FollowUpOpts{Color} -> BuildFollowUpCommand -> shell line with/without --color
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentrunapi` exports
  `BuildFollowUpCommand` and `FollowUpOpts` with **`Color bool`** (RED until
  implementer adds the field + emission).
- Nested root: `d.DOCTEST_ROOT` is `tests/agentrunapi/follow-up-color`.
- No real agent-run binary / iTerm.

## Steps

1. Root seeds default session / prompt / open profile.
2. Leaf sets `Color` true or false.
3. `Run` calls `BuildFollowUpCommand`; assert token presence.

## Context

- Default: `SessionID=sess-fu-color`, `Prompt=follow-up color`, `AgentRunner=grok-tty`, `Open=true`.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.SessionID == "" {
		req.SessionID = "sess-fu-color"
	}
	if req.Prompt == "" {
		req.Prompt = "follow-up color"
	}
	if req.AgentRunner == "" {
		req.AgentRunner = "grok-tty"
	}
	if !req.Open && !req.Detach {
		req.Open = true
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

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in %q", want, got)
	}
}
```
