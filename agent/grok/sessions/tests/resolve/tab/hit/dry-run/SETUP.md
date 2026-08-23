# Scenario

**Feature**: tab path `--dry-run` prints tab plan

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedTabWindow(req)
	req.Args = []string{"--tab", "2", "--dry-run"}
	return nil
}
```
