---
label: e2e
---

## Expected

- Non-zero exit code.
- Stderr mentions `parse`.
- Stdout does not contain all three success status lines.

## Side Effects

- Ephemeral tty-watch session killed on error path (registry pruned).

## Errors

- Status output parse failure.

## Exit Code

- Non-zero.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertError(t, resp, "parse")
	assertStdoutNotSuccessLines(t, resp)
}
```