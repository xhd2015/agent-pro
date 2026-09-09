# Scenario

**Feature**: `--details` prints session title/cwd/kind on stdout

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedHit(req, fixtureSessionID, pidGrok)
	seedSummary(t, req, fixtureSessionID, fixtureDetailsTitle, fixtureDetailsCWD, "", fixtureDetailsLastActive)
	req.Now = fixtureDetailsNow
	req.Args = []string{"--details"}
	return nil
}
```
