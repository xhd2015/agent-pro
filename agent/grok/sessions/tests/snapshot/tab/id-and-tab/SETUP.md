# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{fixtureSnapshotSessionID, "--tab", "2"}
	return nil
}
```
