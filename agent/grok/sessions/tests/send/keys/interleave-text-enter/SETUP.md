# Scenario

**Feature**: `--up --text pick --enter` then positional tail; NoSubmit for --enter

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectSendSession(t, req)
	addLiveGrok(req, 5001, "ttys148")
	req.ITerm = oneITermTab()
	req.Args = []string{"--up", "--text", "pick", "--enter", "tail", "--session-id", req.SessionID}
	return nil
}
```
