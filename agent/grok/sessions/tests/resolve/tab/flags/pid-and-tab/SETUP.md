# Scenario

**Feature**: `--pid` and `--tab` together error

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--pid", "6000", "--tab", "2"}
	return nil
}
```
