## Expected

- `RunFocus` returns nil error (help success).
- Stdout documents `focus`, `--index`, and `--dry-run` (case-insensitive ok for headers).
- Stdout ends with trailing newline `\n`.

## Errors

- None.

## Exit Code

0 (success / nil error)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	out := resp.Stdout
	if out == "" {
		t.Fatal("help stdout is empty")
	}
	lower := strings.ToLower(out)
	for _, want := range []string{"focus", "--index", "--dry-run"} {
		if !strings.Contains(lower, strings.ToLower(want)) {
			t.Fatalf("help stdout must document %q; got:\n%s", want, out)
		}
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("help stdout must end with trailing newline; got %q", out)
	}
}
```
