## Preconditions
- The leaf tests that `Logf` appends exactly one newline when the message lacks a trailing `\n`.

## Steps
1. Set `req.LogMessage` to `"hello"` (no trailing newline).
2. Call `Logf`, capture stdout.

```go
func Setup(t *testing.T, req *Request) error {
    req.LogMessage = "hello"
    return nil
}
```
