---
label: e2e
---

## Expected

- Exit code 1.
- Stderr (or stdout) mentions that `--new-terminal` requires `--auto-send-or-resume`.

## Exit Code

1

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
	assertExitCode(t, resp, 1)
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertContainsAny(t, combined,
		"requires --auto-send-or-resume",
		"require --auto-send-or-resume",
		"--auto-send-or-resume",
	)
	assertContainsAny(t, combined,
		"new-terminal",
		"--new-terminal",
	)
}
```
