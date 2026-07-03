# Scenario

**Feature**: `tty send` fails when registry entry exists but port is unreachable

```
agent-run tty send session-1 "hello" -> registry with dead port -> error
```

## Steps

1. Registry entry has listen_addr on a closed port.
2. `req.Args` = `["tty", "send", "session-1", "hello"]`.
3. `req.Mode` = `"send-probe"`.
4. Expect exit code 1 or error output.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"tty", "send", req.RegistrySessionID, "hello"}
	req.Mode = "send-probe"
	req.ExecTimeout = 15 * time.Second
	return nil
}
```
