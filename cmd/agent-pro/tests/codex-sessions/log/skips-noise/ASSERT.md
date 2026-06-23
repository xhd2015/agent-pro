## Expected

- `PrintLog` succeeds with empty or whitespace-only output.
- Output does not contain trace labels `RUN`, `ASSISTANT`, `EDIT`, or `REASONING`.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if strings.TrimSpace(resp.Output) != "" {
		t.Fatalf("expected empty log output, got:\n%s", resp.Output)
	}
	for _, label := range []string{"RUN", "ASSISTANT", "EDIT", "REASONING"} {
		assertNotContains(t, resp.Output, label)
	}
}
```