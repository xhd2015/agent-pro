# Scenario

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	addLiveGrokHost(req, 5001, "ttys148", fixtureListLiveSID, "3", 1)
	req.Args = []string{}
	return nil
}
```
