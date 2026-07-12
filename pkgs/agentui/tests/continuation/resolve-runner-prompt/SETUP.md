# Scenario

**Feature**: ResolveRunnerPrompt chooses inject text based on provider resume id

```
# bound resume: provider already has conversation context
resumeID non-empty + prior + newPrompt
  -> ResolveRunnerPrompt -> trimmed newPrompt only

# unbound multi-turn: no native resume — inject history wrap
resumeID empty + prior + newPrompt
  -> ResolveRunnerPrompt -> BuildContinuationPrompt(prior, newPrompt)
```

## Preconditions

- Package `pkgs/agentui` exports
  `ResolveRunnerPrompt(resumeID, newPrompt string, prior []types.AgentEvent) string`
  (RED until implementer adds it).
- Tests are pure Go (no `AGENT_RUN_HOME`, no CLI).
- Nested root: parent continuation helpers are **not** inherited — helpers defined here.

## Steps

1. Root `Setup` ensures `req.NewPrompt` default when leaf omits it.
2. Leaf `Setup` sets `req.ResumeID`, `req.PriorEvents`, and `req.NewPrompt`.
3. `Run` calls `agentui.ResolveRunnerPrompt`.
4. Leaf `Assert` checks equality with raw new prompt and/or history prefix presence.

## Context

- Resume id examples: `grok-sess-abc`, codex thread ids, any non-empty `RunnerSessionID`.
- Empty resume id is unbound; non-empty means native resume will restore provider context.
- When resume id is bound, prior transcript is intentionally ignored for the inject string.
- Only `type=message` events with non-empty role/text participate when unbound wrap runs.

```go
import (
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	if strings.TrimSpace(req.NewPrompt) == "" {
		req.NewPrompt = "follow-up please"
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
```
