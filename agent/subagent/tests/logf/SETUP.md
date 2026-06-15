## Preconditions
- The `Logf` function in `github.com/xhd2015/agent-pro/agent/subagent` writes timestamped output to `os.Stdout`.
- Each leaf provides a format message and optional arguments.

## Steps
1. Read `req.LogMessage` as the format string to pass to `subagent.Logf()`.
2. Pass `req.LogArgs` as variadic arguments.
3. Redirect `os.Stdout` to a pipe, call `subagent.Logf(message, args...)`.
4. Restore `os.Stdout`, read captured output, return as `resp.Stdout`.

```go
import (
    "bytes"
    "fmt"
    "os"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.Operation = "logf"
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	message := req.LogMessage

	old := os.Stdout
    r, w, err := os.Pipe()
    if err != nil {
        return nil, fmt.Errorf("create pipe: %w", err)
    }
    os.Stdout = w

    subagent.Logf("%s", fmt.Sprintf(message, req.LogArgs...))

    w.Close()
    os.Stdout = old

    var buf bytes.Buffer
    if _, readErr := buf.ReadFrom(r); readErr != nil {
        return nil, fmt.Errorf("read pipe: %w", readErr)
    }

    return &Response{Stdout: buf.String()}, nil
}
```
