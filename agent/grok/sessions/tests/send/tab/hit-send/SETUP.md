# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeSendSession(t, req.GrokHome, fixtureTabSendSessionID, req.ProjectDir, "tab send")
	seedSendTabWindow(req)
	req.Args = []string{"from-tab", "--tab", "2"}
	return nil
}
```
