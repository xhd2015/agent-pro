# Scenario

**Feature**: screen status shows "idle" when persistent turn is complete

```
agent-run tty status session-1 -> scrollback after prompt+response -> screen: idle
```

## Steps

1. Fake ptywrap sends scrollback containing the prompt and response text.
2. `req.Args` = `["tty", "status", "session-1"]`.
3. `Assert` checks that screen status indicates idle.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FakePTYWrapScrollback = "GROK_TTY_BANNER\r\nGrok › run ls\r\nResponse: file1 file2\r\nGrok › \r\n"
	req.Args = []string{"tty", "status", req.RegistrySessionID}
	return nil
}
```
