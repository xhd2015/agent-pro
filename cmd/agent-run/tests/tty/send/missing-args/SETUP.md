# Scenario

**Feature**: `tty send` without required args returns error

```
agent-run tty send -> missing session-id and message
```

## Steps

1. `req.Args` = `["tty", "send"]`.
2. Exit code 1, stderr mentions missing args.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"tty", "send"}
	return nil
}
```
