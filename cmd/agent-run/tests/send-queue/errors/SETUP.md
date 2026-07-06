# Scenario

**Feature**: send errors surface on stderr without stdout id line

```
missing args / unknown session -> exit 1, no msg_N on stdout
```

## Preconditions

- No live session required for missing-args.
- Isolated `AGENT_RUN_HOME` with empty registry for session-not-found.

## Steps

1. `Setup` sets `req.Operation = "errors"`.
2. Leaf configures invalid invocation.
3. `Run` executes CLI.
4. `Assert` exit 1, stderr error, empty stdout id.

```go
func Setup(t *testing.T, req *Request) error {
	req.Operation = "errors"
	return nil
}
```