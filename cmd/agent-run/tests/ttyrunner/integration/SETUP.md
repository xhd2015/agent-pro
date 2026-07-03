# Scenario

**Feature**: sealed doctest tree regression guard

```
# sealed trees must pass unchanged after ttyrunner extraction
doctest test ./cmd/agent-run/tests/tty/...
doctest test ./cmd/agent-run/tests/grok-tty/...
doctest test ./cmd/agent-run/tests/codex-tty/...
```

## Preconditions

- Sealed trees `tty/`, `grok-tty/`, `codex-tty/` are not modified by this feature.

## Steps

1. Leaf documents regression commands (doc pointer only).
2. CI orchestrator runs sealed trees separately.

```go
func Setup(t *testing.T, req *Request) error {
	req.Operation = "integration"
	return nil
}
```