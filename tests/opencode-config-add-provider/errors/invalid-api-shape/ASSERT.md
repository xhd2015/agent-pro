## Expected

- Non-zero exit code.
- stderr mentions at least one valid api-shape value (`anthropic` or `openai`)
  so the user knows the allowed set.

## Side Effects

- No config file written at the global target.

## Errors

- A validation error reporting `gemini` is not a valid api-shape and naming the
  accepted values.

## Exit Code

- Non-zero.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0\nstderr:%s", resp.Stderr)
	}
	lower := strings.ToLower(resp.Stderr)
	if !strings.Contains(lower, "anthropic") && !strings.Contains(lower, "openai") {
		t.Fatalf("stderr does not mention a valid api-shape value:\n%s", resp.Stderr)
	}
	assertNoConfigFile(t, resp)
}
```
