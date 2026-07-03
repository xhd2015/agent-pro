# Scenario

**Feature**: `agent-run attach` and `agent-run tty attach` produce identical error for missing session

```
agent-run attach bogus-id vs agent-run tty attach bogus-id -> same error message
```

## Steps

1. `req.Args` = `["attach", "session-nonexistent"]`.
2. Exit code 1, error mentions not found.
3. Stderr output should match the tty attach equivalent.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"attach", "session-nonexistent"}
	return nil
}
```
