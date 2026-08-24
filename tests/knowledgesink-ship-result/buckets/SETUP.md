# Scenario

**Feature**: per-bucket disk/git presence rules for add/update/delete

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Leaves override HubMode when git tracking is required.
	if req.HubMode == "" {
		req.HubMode = "plain"
	}
	return nil
}
```
