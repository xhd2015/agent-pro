# Scenario

**Feature**: nested groks — nearest ancestor wins

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// main 3000 -> nearer grok 4242 -> bash 5000 -> start 6000
	req.Procs = []FixtureProc{
		{PID: pidMainGrok, PPID: 1, Cmd: "/usr/local/bin/grok"},
		{PID: pidGrok, PPID: pidMainGrok, Cmd: grokCmdWithIgnoredFlags()},
		{PID: pidBash, PPID: pidGrok, Cmd: "/bin/bash"},
		{PID: pidStart, PPID: pidBash, Cmd: "/usr/local/bin/agent-pro"},
	}
	req.OpenFiles = map[int][]string{
		pidMainGrok: {grokSessionPath(fixtureSessionID)},
		pidGrok:     {grokSessionPath(fixtureNearSessionID)},
	}
	req.PID = pidStart
	req.Args = []string{}
	return nil
}
```
