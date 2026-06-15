## Preconditions
- The `trace-session` subtree tests the `CatchUp` operation mode.

## Steps
1. Set `req.CatchUp = true` in Setup.
2. Each leaf pre-creates a session directory with optional events.jsonl.

```go
func Setup(t *testing.T, req *Request) error {
    req.CatchUp = true
    req.Operation = "trace_session"
    return nil
}
```
