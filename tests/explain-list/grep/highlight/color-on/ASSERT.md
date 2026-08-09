---
label: e2e
---

## Expected

- Exit 0.
- Session listed (title `1 shown of 1`).
- Stdout contains bold-red SGR `\x1b[1;31m` and reset `\x1b[0m`.
- Match span uses original casing: `\x1b[1;31mDocker\x1b[0m` (or equivalent
  contiguous bold-red wrap of `Docker`).
- Q label still colored (cyan `\x1b[1;36m` present).
- Trailing newline.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	assertStdoutEndsWithNewline(t, resp.Stdout)
	assertHasANSI(t, resp.Stdout)

	if !strings.Contains(resp.Stdout, "1 shown of 1") {
		t.Fatalf("expected matching session listed:\n%s", resp.Stdout)
	}

	// Bold red match highlight (locked H1).
	assertContains(t, resp.Stdout, "\x1b[1;31m")
	assertContains(t, resp.Stdout, "\x1b[0m")

	// Original casing inside (or as) the bold-red span.
	if !strings.Contains(resp.Stdout, "\x1b[1;31mDocker\x1b[0m") {
		// Allow extra SGR between open and text only if Docker is still shown;
		// primary lock is contiguous bold-red wrap of original "Docker".
		t.Fatalf("expected bold-red wrap of original casing Docker; got:\n%q", resp.Stdout)
	}

	// Labels still colored when --color is on.
	assertContains(t, resp.Stdout, "\x1b[1;36m") // bold cyan Q
}
```
