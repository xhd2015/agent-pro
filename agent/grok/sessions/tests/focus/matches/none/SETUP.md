# Scenario

**Feature**: live grok PID whose TTY matches no iTerm tab is not found

```
known session + grok pid/tty + empty iTerm list -> not found
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectFocusSession(t, req)
	addLiveGrok(req, 5001, "ttys148")
	req.ITerm = nil
	return nil
}
```
