# Scenario

**Feature**: tree walk keeps only real TTYs from ancestors and descendants

```
# tree: parent tty, serve ??, child tty, unrelated other pid
root=200
  ancestor 100 /dev/ttys148
  root 200 ??
  child 201 /dev/ttys200
  unrelated 999 /dev/ttys999  (not in tree)
-> CollectTTYsFromTree -> contains ttys148 and ttys200; not ??; not ttys999
```

## Preconditions

- Root PID is the serve process (TTY `??`).
- Ancestor and descendant hold real TTYs; blank/`??` skipped.

## Steps

1. Seed a small proc snapshot around root PID 200.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RootPID = 200
	req.Procs = []agentrunapi.ProcRow{
		{PID: 1, PPID: 0, TTY: "??", Cmd: "init"},
		{PID: 100, PPID: 1, TTY: "/dev/ttys148", Cmd: "zsh"},
		{PID: 200, PPID: 100, TTY: "??", Cmd: "agent-run serve"},
		{PID: 201, PPID: 200, TTY: "/dev/ttys200", Cmd: "grok"},
		{PID: 202, PPID: 200, TTY: "  ", Cmd: "blank-tty"},
		{PID: 999, PPID: 1, TTY: "/dev/ttys999", Cmd: "unrelated"},
	}
	return nil
}
```
