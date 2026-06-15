## Preconditions
- The leaf tests that `Logf` correctly resolves format verbs.

## Steps
1. Set `req.LogMessage` to `"value=%s count=%d"` with format verbs.
2. Set `req.LogArgs` to `["foo", "42"]`.
3. Call `Logf`, capture stdout.

```go
func Setup(t *testing.T, req *Request) error {
    req.LogMessage = "value=%s count=%s"
    req.LogArgs = []any{"foo", "42"}
    return nil
}
```
