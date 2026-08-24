# Scenario

**Feature**: multiple hosting tabs need --index

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectSnapshotSession(t, req)
	addLiveGrok(req, 5001, "ttys148")
	addLiveGrok(req, 5002, "ttys149")
	req.ITerm = twoITermTabs()
	return nil
}
```
