# Scenario

**Feature**: two grok sessions on target tab → refuse

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedTabMultiGrok(req)
	req.Args = []string{"--tab", "2"}
	return nil
}
```
