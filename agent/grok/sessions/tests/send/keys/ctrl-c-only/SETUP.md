# Scenario

**Feature**: `--ctrl-c` alone sends Ctrl-C with NoCtrlU+NoSubmit

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectSendSession(t, req)
	addLiveGrok(req, 5001, "ttys148")
	req.ITerm = oneITermTab()
	req.Args = []string{"--ctrl-c", "--session-id", req.SessionID}
	return nil
}
```
