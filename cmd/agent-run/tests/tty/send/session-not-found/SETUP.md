# Scenario

**Feature**: `tty send` with unknown session id returns error

```
agent-run tty send bogus-id "hi" -> registry file not found -> error
```

## Steps

1. `req.Args` = `["tty", "send", "session-nonexistent", "hello"]` (no registry file).
2. Exit code 1, stderr mentions session not found.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"tty", "send", "session-nonexistent", "hello world"}
	return nil
}
```
