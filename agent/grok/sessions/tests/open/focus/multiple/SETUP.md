# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectOpenSession(t, req)
	addLiveGrok(req, 5001, "ttys148")
	addLiveGrok(req, 5002, "ttys149")
	req.ITerm = twoITermTabs()
	return nil
}
```
