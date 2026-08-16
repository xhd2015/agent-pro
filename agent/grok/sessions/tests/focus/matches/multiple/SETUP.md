# Scenario

**Feature**: several iTerm tabs host the same Grok session

```
two grok pids on different ttys + two iTerm tabs
  -> require --index / select / reject out of range
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectFocusSession(t, req)
	addLiveGrok(req, 5001, "ttys148")
	addLiveGrok(req, 5002, "ttys149")
	req.ITerm = twoITermTabs()
	return nil
}
```
