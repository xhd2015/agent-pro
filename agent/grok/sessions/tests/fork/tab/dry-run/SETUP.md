# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeForkSession(t, req.GrokHome, fixtureTabSessionID, req.ProjectDir, "tab fixture")
	seedTabWindow(req)
	req.Args = []string{"--tab", "2", "--dry-run"}
	return nil
}
```
