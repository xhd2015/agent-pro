# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeListLiveSession(t, req.GrokHome, fixtureListLiveSID, fixtureListLiveDiskCWD, "json-title")
	addLiveGrokHost(req, 5001, "ttys148", fixtureListLiveSID, "3", 1)
	req.DiskCwd = true
	req.CwdBySession = nil
	req.Args = []string{"--json"}
	return nil
}
```
