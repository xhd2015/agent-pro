# Scenario

**Feature**: extra positional argument is rejected

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedHit(req, fixtureSessionID, pidGrok)
	req.Args = []string{"extra"}
	return nil
}
```
