# Scenario

**Feature**: `--json --details` includes session identity fields

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedHit(req, fixtureSessionID, pidGrok)
	seedSummary(t, req, fixtureSessionID, fixtureDetailsTitle, fixtureDetailsCWD, "subagent", fixtureDetailsLastActive)
	req.Now = fixtureDetailsNow
	req.Args = []string{"--json", "--details"}
	return nil
}
```
