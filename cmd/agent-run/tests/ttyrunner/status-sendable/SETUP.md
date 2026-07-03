# Scenario

**Feature**: tty status sendable fields from provider CheckWritable

```
agent-run tty status --json -> fetch scrollback -> CheckWritable -> sendable fields
```

## Preconditions

- Fake ptywrap server serves deterministic scrollback per leaf.
- Registry entry points at the fake server listen addr.

## Steps

1. Leaf sets scrollback fixture and registry dir (grok vs codex).
2. `Run` executes `agent-run tty status <id> --json`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Operation = "status-sendable"
	req.StartFakePTYWrap = true
	req.RegistrySessionID = "session-1"
	return nil
}
```
