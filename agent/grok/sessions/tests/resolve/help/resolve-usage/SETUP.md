# Scenario

**Feature**: `resolve -h` prints usage

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"-h"}
	return nil
}
```
