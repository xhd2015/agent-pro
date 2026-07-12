## Expected

- Exit code 0.
- No `runner/id` compound session references in stdout.
- Stdout ends with trailing `\n` (empty store may be header-only or blank line).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if !strings.HasSuffix(resp.Stdout, "\n") && resp.Stdout != "" {
		// empty stdout is acceptable only if we still require trailing newline;
		// requirement: user-facing stdout ends with \n after last content line.
		// Prefer non-empty with final \n (header) or just "\n".
		t.Fatalf("expected trailing newline (or empty with newline), got %q", resp.Stdout)
	}
	if resp.Stdout == "" {
		t.Fatal("expected at least a trailing newline on stdout")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("expected trailing newline, got %q", resp.Stdout)
	}
	for _, line := range strings.Split(strings.TrimRight(resp.Stdout, "\n"), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		// header or data: first column must not look like runner/id
		if strings.Contains(fields[0], "/") {
			t.Fatalf("unexpected compound session ref in list line %q", line)
		}
	}
}
```
