# Scenario

**Feature**: `tty status` marks session as TCP unreachable when port is closed

```
agent-run tty status session-1 -> registry file with dead port -> tcp reachable: false
```

## Steps

1. Mock registry entry has listen_addr on a port that has no server listening.
2. `req.Args` = `["tty", "status", "session-1"]`.
3. Output should show tcp reachable as false/unreachable.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"tty", "status", req.RegistrySessionID}
	return nil
}
```
