## Expected Output

Stderr contains a prefixed session id; stdout does not leak the registry prefix.

```
<contains>
grok-tty: session-
</contains>
```

## Expected

- Exit code 0.
- Stderr matches `grok-tty: session-\d+`.
- Stdout does not contain substring `grok-tty:`.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assert.Output(t, resp.Stderr, `
<contains>
<regex>grok-tty:\s*session-\d+</regex>
</contains>`)
	if strings.Contains(resp.Stdout, "grok-tty:") {
		t.Fatalf("session id prefix must not appear on stdout:\n%s", resp.Stdout)
	}
}
```