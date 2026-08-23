# Scenario

**Feature**: `--tab next` on last tab errors (no wrap)

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedTabCurrentIsLast(req)
	req.Args = []string{"--tab", "next"}
	return nil
}
```
