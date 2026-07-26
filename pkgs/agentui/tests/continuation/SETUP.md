# Scenario

**Feature**: continuation prompt construction from prior timeline messages

```
types.AgentEvent history + new user text -> BuildContinuationPrompt -> runner-ready prompt
```

## Preconditions

- Package `pkgs/agentui` exports `BuildContinuationPrompt(events []types.AgentEvent, newPrompt string) string`.
- Tests are pure Go (no `AGENT_RUN_HOME`, no CLI).
- Resume-id gate coverage is a nested tree under `resolve-runner-prompt/` (separate `DOCTEST.md`).

## Steps

1. Root `Setup` ensures `req.NewPrompt` default when leaf omits it.
2. Leaf `Setup` sets `req.PriorEvents` and `req.NewPrompt`.
3. `Run` calls `agentui.BuildContinuationPrompt`.
4. Leaf `Assert` checks substring, absence of duplicate blocks, or ordering.

## Context

- Only `type=message` events with non-empty `role` and `text` participate in history.
- Non-message events (tool, done, think) are ignored for v1 prefix building.

```go
import (
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if strings.TrimSpace(req.NewPrompt) == "" {
		req.NewPrompt = "what did I ask?"
	}
	return nil
}

func msgEvent(role, text string) types.AgentEvent {
	return types.AgentEvent{
		Type: types.ActionMessage,
		Role: role,
		Text: text,
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected built prompt to contain %q, got:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("expected built prompt NOT to contain %q, got:\n%s", want, got)
	}
}

func indexOf(t *testing.T, s, sub string) int {
	t.Helper()
	i := strings.Index(s, sub)
	if i < 0 {
		t.Fatalf("expected %q in:\n%s", sub, s)
	}
	return i
}
```
