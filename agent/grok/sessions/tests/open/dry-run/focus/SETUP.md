# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectOpenSession(t, req)
	addLiveGrok(req, 5001, "ttys148")
	req.ITerm = oneITermTab()
	req.Args = []string{req.SessionID, "--dry-run"}
	return nil
}
```
