# Scenario

**Feature**: `--pid` selects an alternate ancestor chain

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Procs = []FixtureProc{
		{PID: pidGrok, PPID: 1, Cmd: grokCmdWithIgnoredFlags()},
		{PID: pidBash, PPID: pidGrok, Cmd: "/bin/bash"},
		{PID: pidStart, PPID: pidBash, Cmd: "/usr/local/bin/agent-pro"},
		{PID: pidAltGrok, PPID: 1, Cmd: "/usr/local/bin/grok"},
		{PID: pidAltStart, PPID: pidAltGrok, Cmd: "/bin/bash"},
	}
	req.OpenFiles = map[int][]string{
		pidGrok:    {grokSessionPath(fixtureSessionID)},
		pidAltGrok: {grokSessionPath(fixtureAltSessionID)},
	}
	req.PID = pidStart // default would hit fixtureSessionID
	req.Args = []string{"--pid", "7100"}
	return nil
}
```
