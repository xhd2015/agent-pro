# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeOpenSession(t, req.GrokHome, req.SessionID, "", "empty cwd")
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	req.ITerm = nil
	return nil
}
```
