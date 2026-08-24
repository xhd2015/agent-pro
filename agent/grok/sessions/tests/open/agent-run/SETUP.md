# Scenario

**Feature**: open prefers agent-run when Grok id is managed

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.SessionID = fixtureOpenSessionID
	req.Args = []string{req.SessionID}
	return nil
}
```
