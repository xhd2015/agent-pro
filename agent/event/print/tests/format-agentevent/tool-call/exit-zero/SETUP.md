## Preconditions
- A tool_call event with exit_code 0 (success).

## Steps
1. Set `req.Output = "ok"`.
2. Set `req.ExitCode` to a pointer to 0 (success).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	exitZero := 0
	req.Output = "ok"
	req.ExitCode = &exitZero
	return nil
}
```
