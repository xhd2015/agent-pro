# Scenario

Two live grok hosts: shared probes must not re-run `ps` / `lsof` / iTerm list
per session id.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	addLiveGrokHost(req, 5001, "ttys148", fixtureListLiveSID, "3", 1)
	addLiveGrokHost(req, 5002, "ttys149", fixtureListLiveSID2, "4", 1)
	req.Args = nil
	return nil
}
```
