# Scenario

**Feature**: `--json` prints detail object

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedHit(req, fixtureSessionID, pidGrok)
	req.Args = []string{"--json"}
	return nil
}
```
