## Expected

- Exit 0.
- Title includes `10 shown of 12` and `limit 10` (effective default, not limit 0).
- Newest 10 present; oldest 2 absent.
- Trailing newline; no ANSI.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- 0.

```go
import (
	"fmt"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	assertStdoutEndsWithNewline(t, resp.Stdout)
	assertNoANSI(t, resp.Stdout)

	if !strings.Contains(resp.Stdout, "10 shown of 12") || !strings.Contains(resp.Stdout, "limit 10") {
		t.Fatalf(" --limit 0 must behave as default limit 10:\n%s", resp.Stdout)
	}
	// Must not advertise limit 0.
	if strings.Contains(resp.Stdout, "limit 0") {
		t.Fatalf("title must not report limit 0 after defaulting:\n%s", resp.Stdout)
	}
	for i := 2; i <= 11; i++ {
		assertContains(t, resp.Stdout, fmt.Sprintf("question-%02d", i))
	}
	assertNotContains(t, resp.Stdout, "question-00")
	assertNotContains(t, resp.Stdout, "question-01")
}
```
