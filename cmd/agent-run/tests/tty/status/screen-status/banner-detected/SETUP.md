# Scenario

**Feature**: screen status shows "banner" when scrollback contains the TUI banner marker

```
agent-run tty status session-1 -> scrollback has GROK_TTY_BANNER -> screen: banner
```

## Steps

1. Fake ptywrap sends scrollback containing `GROK_TTY_BANNER`.
2. `req.Args` = `["tty", "status", "session-1"]`.
3. `Assert` checks that screen status indicates banner detected.

```go
func Setup(t *testing.T, req *Request) error {
	req.FakePTYWrapScrollback = "GROK_TTY_BANNER\r\nGrok > \r\n"
	req.Args = []string{"tty", "status", req.RegistrySessionID}
	return nil
}
```
