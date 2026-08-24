# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = "019f283a-dead-7ead-dead-deaddeaddead"
	req.Args = []string{req.SessionID}
	return nil
}
```
