# Scenario

**Feature**: key / --text sequence flags compose an ordered send payload

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.SessionID = fixtureSendSessionID
	return nil
}
```
