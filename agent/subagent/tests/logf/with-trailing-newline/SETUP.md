## Preconditions
- The leaf tests that `Logf` does not double the newline when the message already ends with `\n`.

## Steps
1. Set `req.LogMessage` to `"hello\n"` (with trailing newline).
2. Call `Logf`, capture stdout.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.LogMessage = "hello\n"
    return nil
}
```
