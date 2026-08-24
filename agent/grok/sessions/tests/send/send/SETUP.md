# Scenario

**Feature**: send via --session-id to a live host

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.SessionID = fixtureSendSessionID
	return nil
}
```
