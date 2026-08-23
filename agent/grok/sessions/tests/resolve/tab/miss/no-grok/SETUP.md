# Scenario

**Feature**: `--tab 3` bash-only tab → no grok session

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedTabWindow(req)
	req.Args = []string{"--tab", "3"}
	return nil
}
```
