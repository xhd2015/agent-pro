# Scenario

**Feature**: --dry-run plan without Contents

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureSnapshotSessionID
	writeProjectSnapshotSession(t, req)
	addLiveGrok(req, 5001, "ttys148")
	req.ITerm = oneITermTab()
	req.Args = []string{req.SessionID, "--dry-run"}
	return nil
}
```
