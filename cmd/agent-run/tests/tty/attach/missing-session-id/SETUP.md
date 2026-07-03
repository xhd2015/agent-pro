# Scenario

**Feature**: `tty attach` without session id returns error

```
agent-run tty attach -> missing session id error + usage hint
```

## Steps

1. `req.Args` = `["tty", "attach"]`.
2. Exit code 1, stderr mentions missing session id.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"tty", "attach"}
	return nil
}
```
