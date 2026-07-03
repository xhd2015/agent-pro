# Scenario

**Feature**: `tty status <session-id>` with no matching registry returns error

```
agent-run tty status bogus-id -> registry file not found -> error
```

## Steps

1. `req.Args` set to `["tty", "status", "bogus-id"]` (no registry file written).
2. `Run` executes the command.
3. `Assert` checks exit code 1 and session-not-found error.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"tty", "status", "session-nonexistent"}
	return nil
}
```
