# Scenario

**Feature**: idle scrollback reports sendable yes

```
scrollback with prompt -> tty status --json -> sendable: true
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "idle-screen-sendable-yes"
	req.RegistryDir = "grok-tty-registry"
	// Modern boxed composer (section judge). Legacy "Grok ›" / "Response:" is no longer writable-idle.
	req.FakePTYWrapScrollback = "" +
		"GROK_TTY_BANNER\n" +
		" ⎇ master worktree ~/.wrk/… 1K / 10K\n" +
		"    Worked for 1.0s                                        stop  [hooks: 1]\n" +
		" ╭--------------------------------------------------------------------------╮\n" +
		" │ ❯                                                                        │\n" +
		" ╰----------------------------------------- Grok 4.5 (high) · always-approve -╯\n" +
		" Shift+Tab:mode  │  Ctrl+.:shortcuts\n"
	return nil
}
```
