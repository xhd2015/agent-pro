# Scenario

**Feature**: parent + child subagent on same tab → resolve parent

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedTabParentSubagent(req)
	req.Args = []string{"--tab", "2"}
	return nil
}
```
