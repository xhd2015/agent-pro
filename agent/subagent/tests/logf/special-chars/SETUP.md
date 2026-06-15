## Preconditions
- The leaf tests that `Logf` preserves multiline content and special characters.

## Steps
1. Set `req.LogMessage` to a multiline string with special characters: `"line1\nline2\tindented\nline3\n"`.
2. Call `Logf`, capture stdout.

```go
func Setup(t *testing.T, req *Request) error {
    req.LogMessage = "line1\nline2\tindented\nline3\n"
    return nil
}
```
