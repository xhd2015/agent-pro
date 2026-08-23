# Scenario

**Feature**: `--tab-index 1` (0-based) resolves second tab

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedTabWindow(req)
	req.Args = []string{"--tab-index", "1"}
	return nil
}
```
