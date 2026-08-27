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
	// Modern boxed composer (section judge). Legacy "Grok ›" is no longer writable-idle.
	req.FakePTYWrapScrollback = "" +
		"GROK_TTY_BANNER\r\n" +
		"run ls\r\n" +
		"Response: file1 file2\r\n" +
		" ⎇ master worktree ~/.wrk/… 1K / 10K\r\n" +
		"    Worked for 1.0s                                        stop  [hooks: 1]\r\n" +
		" ╭--------------------------------------------------------------------------╮\r\n" +
		" │ ❯                                                                        │\r\n" +
		" ╰----------------------------------------- Grok 4.5 (high) · always-approve -╯\r\n" +
		" Shift+Tab:mode  │  Ctrl+.:shortcuts\r\n"
	req.Args = []string{"tty", "status", req.RegistrySessionID}
	return nil
}
```
