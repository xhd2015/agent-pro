---
label: e2e
---

## Expected

- Non-zero exit code.
- stderr mentions `--base-url`.

## Side Effects

- No config file written at the global target.

## Errors

- A validation error naming the missing `--base-url` flag.

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
	assertError(t, resp, "--base-url")
	assertNoConfigFile(t, resp)
}
```
