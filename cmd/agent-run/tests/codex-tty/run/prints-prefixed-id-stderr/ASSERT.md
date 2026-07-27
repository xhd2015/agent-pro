## Expected Output

Stderr contains a prefixed session id; stdout does not leak the registry prefix.

```
<contains>
codex-tty: session-
</contains>
```

## Expected

- Exit code 0.
- Stderr matches `codex-tty: session-\d+`.
- Stdout does not contain substring `codex-tty:`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assert.Output(t, resp.Stderr, `` +
`<contains>
<regex>codex-tty:\s*session-\d+</regex>
</contains>`)
	if strings.Contains(resp.Stdout, "codex-tty:") {
		t.Fatalf("session id prefix must not appear on stdout:\n%s", resp.Stdout)
	}
}
```
