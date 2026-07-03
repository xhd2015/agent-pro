# Scenario

**Feature**: `tty status --json` outputs valid JSON with all registry fields

```
agent-run tty status --json session-1 -> valid JSON object with pid, port, tty_type, session_id, ...
```

## Steps

1. `req.Args` = `["tty", "status", "--json", "session-1"]`.
2. `req.Mode` = `"status-json"` for JSON parsing.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"tty", "status", "--json", req.RegistrySessionID}
	req.Mode = "status-json"
	return nil
}
```
