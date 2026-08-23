# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeForkSession(t, req.GrokHome, fixtureForkSessionID, req.ProjectDir, "fork fixture")
	req.Args = []string{fixtureForkSessionID, "--dry-run"}
	return nil
}
```
