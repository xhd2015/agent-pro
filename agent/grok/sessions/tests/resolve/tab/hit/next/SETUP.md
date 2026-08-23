# Scenario

**Feature**: `--tab next` from first tab selects second

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedTabWindow(req)
	req.Args = []string{"--tab", "next"}
	return nil
}
```
