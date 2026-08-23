# Scenario

**Feature**: `--tab 0` invalid (1-based)

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--tab", "0"}
	return nil
}
```
