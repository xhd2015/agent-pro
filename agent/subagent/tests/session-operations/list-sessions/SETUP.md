## Preconditions
- The `list-sessions` subtree tests the `ListSessions` operation mode.

## Steps
1. Set `req.ListSessions = true` in Setup.
2. Each leaf pre-creates session directories (or none) under the configured base.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.ListSessions = true
    req.Operation = "list_sessions"
    return nil
}
```
