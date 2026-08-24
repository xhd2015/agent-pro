# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectSendSession(t, req)
	addLiveGrok(req, 5001, "ttys148")
	req.ITerm = oneITermTab()
	req.Args = []string{"partial", "--session-id", req.SessionID, "--no-submit", "--focus", "--no-ctrl-u"}
	return nil
}
```
