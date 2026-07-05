# Scenario

**Bug**: `agent-run send <session-id> "message"` is an alias for `agent-run tty send <session-id> "message"`

```
agent-run send <id> "msg" -> delegates to same logic as tty send -> same error output
agent-run tty send <id> "msg" -> delegates to same logic -> same error output
```

## Preconditions

- Both `send` and `tty send` delegate to the same implementation.
- An error produced by one must match the error from the other when given the same input.

## Steps

1. `Setup` configures a bad session id (no registry entry) or missing args.
2. Paired leaves run `send` and compare behavior to the equivalent `tty send` case.

```go
func Setup(t *testing.T, req *Request) error {
	req.RegistryDir = "grok-tty-registry"
	req.RegistrySessionID = "session-nonexistent"
	return nil
}
```