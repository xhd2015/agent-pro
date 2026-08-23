# Scenario

**Feature**: grok on different PTY than iTerm tab still resolves via ancestor TTY tree

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedTabWrappedGrok(req)
	req.Args = []string{"--tab", "2"}
	return nil
}
```
