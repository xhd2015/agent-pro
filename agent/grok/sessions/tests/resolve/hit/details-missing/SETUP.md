# Scenario

**Feature**: `--details` soft-succeeds when summary is missing

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedHit(req, fixtureSessionID, pidGrok)
	// No seedSummary — Find misses; still print bare id.
	req.Args = []string{"--details"}
	return nil
}
```
