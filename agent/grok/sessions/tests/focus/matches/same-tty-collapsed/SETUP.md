# Scenario

**Feature**: two grok PIDs on one TTY still focus one tab

```
two grok pids, same tty, one iTerm tab -> focus without --index
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectFocusSession(t, req)
	addLiveGrok(req, 5001, "ttys148")
	addLiveGrok(req, 5002, "ttys148")
	req.ITerm = oneITermTab()
	return nil
}
```
