# Scenario

**Feature**: `--tab left` from second tab selects first

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedTabLeftHit(req)
	req.Args = []string{"--tab", "left"}
	return nil
}
```
