## Preconditions
- A tool_call event with non-zero exit_code (failure).

## Steps
1. Override `req.Tool = "read"` (to test READ icon).
2. Set `req.Output = "file not found"`.
3. Set `req.ExitCode` to a pointer to 1 (error).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	exitErr := 1
	req.Tool = "read"
	req.Output = "file not found"
	req.ExitCode = &exitErr
	return nil
}
```
