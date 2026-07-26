## Preconditions
- The `show-status` subtree tests the `Status` operation mode.

## Steps
1. Set `req.Status = true` in Setup.
2. Each leaf pre-creates a session directory with meta.json and optionally events/pids.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Status = true
    req.Operation = "show_status"
    return nil
}
```
