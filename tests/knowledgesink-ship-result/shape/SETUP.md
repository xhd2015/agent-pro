# Scenario

**Feature**: wire shape of `git_commit_files` (object vs legacy array)

## Preconditions

- Plain hub with one existing file is enough for success leaf.
- Legacy array leaf does not need a valid path set to exist.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.HubMode = "plain"
	return nil
}
```
