# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeOpenSession(t, req.GrokHome, fixtureTabOpenSessionID, req.ProjectDir, "tab open")
	seedOpenTabWindow(req)
	req.Args = []string{"--tab", "2", "--dry-run"}
	return nil
}
```
