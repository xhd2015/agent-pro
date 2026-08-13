# Scenario

**Feature**: nested groks — Mode A forks the nearest session, not the topmost

```
# main grok 3000 (main session)
#   sub grok 4242 (fixture session)
#     bash → start 6000
fork.Main([]) -> --session-id fixture (nearest), not main
```

## Steps

1. Insert main grok above 4242; seed both sessions.
2. Bare args.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainDir := seedSession(t, req.GrokHome, fixtureMainSessionID, req.Workspace)
	req.Procs = []FixtureProc{
		{PID: pidMainGrok, PPID: 1, Cmd: "/usr/local/bin/grok"},
		{PID: pidGrok, PPID: pidMainGrok, Cmd: grokCmdWithIgnoredFlags()},
		{PID: pidBash, PPID: pidGrok, Cmd: "/bin/bash"},
		{PID: pidStart, PPID: pidBash, Cmd: "/usr/local/bin/grok-fork"},
	}
	req.OpenFiles[pidMainGrok] = []string{lsofPath(mainDir)}
	req.Args = []string{}
	return nil
}
```
