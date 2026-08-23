# Scenario

**Feature**: start pid is the grok runner itself

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Procs = []FixtureProc{
		{PID: pidGrok, PPID: 1, Cmd: grokCmdWithIgnoredFlags()},
	}
	req.OpenFiles = map[int][]string{
		pidGrok: {grokSessionPath(fixtureSessionID)},
	}
	req.PID = pidGrok
	req.Args = []string{"--json"}
	return nil
}
```
