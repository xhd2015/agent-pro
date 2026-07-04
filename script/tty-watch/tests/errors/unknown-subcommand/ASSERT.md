## Expected

- Non-zero exit code.
- Combined output mentions unknown subcommand (or usage hint).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertNonZeroExit(t, resp)
	lower := strings.ToLower(resp.Combined)
	if !strings.Contains(lower, "unknown") && !strings.Contains(lower, "subcommand") && !strings.Contains(lower, "usage") {
		t.Fatalf("expected unknown subcommand error, got combined %q", resp.Combined)
	}
}
```