# Scenario

**Feature**: `--dry-run` prints plan with `[dry-run]` prefix

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedHit(req, fixtureSessionID, pidGrok)
	req.Args = []string{"--dry-run"}
	return nil
}
```
