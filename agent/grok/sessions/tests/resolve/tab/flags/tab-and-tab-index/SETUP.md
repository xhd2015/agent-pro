# Scenario

**Feature**: `--tab` and `--tab-index` together error

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--tab", "2", "--tab-index", "1"}
	return nil
}
```
