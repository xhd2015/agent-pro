## Preconditions
- The leaf tests that `Logf` with an empty message still produces a timestamp and newline.

## Steps
1. Set `req.LogMessage` to `""` (empty string).
2. Call `Logf`, capture stdout.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.LogMessage = ""
    return nil
}
```
