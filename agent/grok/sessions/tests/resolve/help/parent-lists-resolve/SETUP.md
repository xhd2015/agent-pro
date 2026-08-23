# Scenario

**Feature**: parent `grok session` help lists resolve

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ParentHelp = true
	return nil
}
```
