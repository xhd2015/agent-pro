# Scenario

**Feature**: no grok on the PPID chain → error, no open

```
# bash → grok-fork only
fork.Main([]) -> error "no ancestor grok"
```

## Steps

1. Replace procs with a bash-only chain (no grok).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Procs = []FixtureProc{
		{PID: 1, PPID: 0, Cmd: "/sbin/init"},
		{PID: pidBash, PPID: 1, Cmd: "/bin/bash"},
		{PID: pidStart, PPID: pidBash, Cmd: "/usr/local/bin/grok-fork"},
	}
	req.OpenFiles = map[int][]string{}
	req.Args = []string{}
	return nil
}
```
