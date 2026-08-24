# Scenario

**Feature**: --dry-run resolve only

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.SessionID = fixtureSendSessionID
	return nil
}
```
