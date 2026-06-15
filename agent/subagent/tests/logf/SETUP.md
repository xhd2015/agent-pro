## Preconditions
- The `Logf` function in `github.com/xhd2015/agent-pro/agent/subagent` writes timestamped output to `os.Stdout`.
- Each leaf provides a format message and optional arguments.

## Steps
1. Read `req.LogMessage` as the format string to pass to `subagent.Logf()`.
2. Pass `req.LogArgs` as variadic arguments.
3. Redirect `os.Stdout` to a pipe, call `subagent.Logf(message, args...)`.
4. Restore `os.Stdout`, read captured output, return as `resp.Stdout`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Operation = "logf"
    return nil
}
```
