# Scenario

**Feature**: --open resume-then-send

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.SessionID = fixtureSendSessionID
	return nil
}
```
