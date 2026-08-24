# Scenario

**Feature**: capture via positional Grok session id

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.SessionID = fixtureSnapshotSessionID
	req.Args = []string{req.SessionID}
	return nil
}
```
