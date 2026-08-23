# Scenario

**Feature**: `--tab left` on first tab errors (no wrap)

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedTabWindow(req)
	req.Args = []string{"--tab", "left"}
	return nil
}
```
