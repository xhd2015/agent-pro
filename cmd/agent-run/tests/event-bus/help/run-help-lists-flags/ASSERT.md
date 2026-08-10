## Expected

- `RunHelpText()` returns non-empty help body (exit 0).
- Text contains `--event-bus-url`.
- Text contains `--event-bus-token`.
- Text ends with trailing newline `\n` (CLI stdout contract).

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code: got %d, want 0", resp.ExitCode)
	}
	if strings.TrimSpace(resp.Stdout) == "" {
		t.Fatal("RunHelpText returned empty help")
	}
	for _, flag := range []string{"--event-bus-url", "--event-bus-token"} {
		if !strings.Contains(resp.Stdout, flag) {
			t.Fatalf("run help must document %s; text:\n%s", flag, resp.Stdout)
		}
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("run help text must end with trailing newline")
	}
}
```
