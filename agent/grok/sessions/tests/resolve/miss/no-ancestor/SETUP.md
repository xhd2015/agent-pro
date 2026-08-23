# Scenario

**Feature**: no grok on ancestor chain (descendant decoy ignored)

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Procs = []FixtureProc{
		{PID: 1, PPID: 0, Cmd: "/sbin/launchd"},
		{PID: pidBash, PPID: 1, Cmd: "/bin/bash"},
		{PID: pidStart, PPID: pidBash, Cmd: "/usr/local/bin/agent-pro"},
		// decoy descendant grok must not be used
		{PID: 9000, PPID: pidStart, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		9000: {grokSessionPath(fixtureSessionID)},
	}
	req.PID = pidStart
	req.Args = []string{}
	return nil
}
```
