# Scenario

**Feature**: `-v` prints details on stderr

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedHit(req, fixtureSessionID, pidGrok)
	req.Args = []string{"-v"}
	return nil
}
```
