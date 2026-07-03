# Scenario

**Feature**: test-only stub-tty full alt-screen runner with scenario DSL

```
AGENT_RUN_ENABLE_STUB_TTY=1 + AGENT_RUN_STUB_TTY_SCENARIO -> stub-tty run -> registry + events + tty.json
```

## Preconditions

- `stub-tty` registered when `AGENT_RUN_ENABLE_STUB_TTY=1`.
- Scenario JSON controls banner delay, screen frames, llm_events, screen_status.

## Steps

1. Leaf writes or selects scenario JSON.
2. `Run` executes `agent-run run --agent-runner stub-tty` or background keep-tty session.

```go
func Setup(t *testing.T, req *Request) error {
	req.Operation = "stub-tty"
	req.EnableStubTTY = true
	return nil
}
```
