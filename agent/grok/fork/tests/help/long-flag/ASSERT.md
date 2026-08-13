## Expected Output

Usage mentions `--session-id`, `--dir`, `--pid`, `--dry-run`, `--color`,
`--no-color`. Does not mention `-n` or `--new-terminal`. Ends with a newline.

## Expected

- `Main` returns nil; exit 0.
- Stdout contains the required flags.
- No `-n` / `--new-terminal`.
- No ANSI (help is plain).
- No OpenInNewTerminal / RunForeground.

## Side Effects

- None.

## Errors

- None.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainOK(t, resp)
	if resp.Stdout == "" {
		t.Fatal("empty help stdout")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with newline: %q", resp.Stdout)
	}
	helpMentions(t, resp.Stdout)
	assertNoANSI(t, resp.Stdout, "help stdout")
	assertNoOpen(t, resp)
	assertNoForeground(t, resp)
}
```
