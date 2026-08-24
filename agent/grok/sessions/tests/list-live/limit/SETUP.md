# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	addLiveGrokHost(req, 5001, "ttys148", fixtureListLiveSID, "3", 1)
	addLiveGrokHost(req, 5002, "ttys149", fixtureListLiveSID2, "4", 1)
	req.Args = []string{"--limit", "1"}
	return nil
}
```
