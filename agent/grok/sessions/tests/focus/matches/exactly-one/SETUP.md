# Scenario

**Feature**: the sole hosting tab is focused

```
known session + one grok tty + one iTerm tab -> focus that tab
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectFocusSession(t, req)
	addLiveGrok(req, 5001, "ttys148")
	req.ITerm = oneITermTab()
	return nil
}
```
