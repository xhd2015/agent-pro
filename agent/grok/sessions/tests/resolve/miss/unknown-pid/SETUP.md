# Scenario

**Feature**: start pid absent from process snapshot

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Procs = defaultAncestorChain()
	req.PID = 999999
	req.Args = []string{"--pid", "999999"}
	return nil
}
```
