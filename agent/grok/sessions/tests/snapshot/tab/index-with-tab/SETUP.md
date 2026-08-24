# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeSnapshotSession(t, req.GrokHome, fixtureTabSnapshotSessionID, req.ProjectDir, "tab snapshot")
	seedSnapshotTabWindow(req)
	req.Args = []string{"--tab", "2", "--index", "0"}
	return nil
}
```
