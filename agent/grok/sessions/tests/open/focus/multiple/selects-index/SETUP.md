# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{req.SessionID, "--index", "1"}
	return nil
}
```
