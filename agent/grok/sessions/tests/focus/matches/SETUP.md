# Scenario

**Feature**: focus resolves a known Grok session to the tab that hosts its live process

```
writeFocusSession
  -> focus <id>
  -> live grok PID + TTY -> 0 / 1 / many iTerm tabs
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.SessionID = fixtureFocusSessionID
	req.Args = []string{req.SessionID}
	return nil
}
```
