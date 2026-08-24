# Scenario

**Feature**: cross-bucket path identity conflicts

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.HubMode = "plain"
	return nil
}
```
