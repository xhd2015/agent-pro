# Scenario

**Feature**: `--up --up --enter` key-only menu navigation

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectSendSession(t, req)
	addLiveGrok(req, 5001, "ttys148")
	req.ITerm = oneITermTab()
	req.Args = []string{"--up", "--up", "--enter", "--session-id", req.SessionID}
	return nil
}
```
