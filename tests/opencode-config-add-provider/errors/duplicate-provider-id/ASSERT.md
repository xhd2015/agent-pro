---
label: e2e
---

## Expected

- Non-zero exit code.
- stderr mentions the colliding id `dup`.

## Side Effects

- The config file is **unchanged** — byte-for-byte equal to its pre-command
  content. No new provider is added, no temp+rename lands.

## Errors

- A duplicate-provider-id error naming `dup`.

## Exit Code

- Non-zero.

```go
import (
	"bytes"
	"os"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertError(t, resp, "dup")

	after, err := os.ReadFile(resp.ConfigPath)
	if err != nil {
		t.Fatalf("config file missing after failed command: %v", err)
	}
	if !bytes.Equal(after, req.Snapshot) {
		t.Fatalf("config file changed after duplicate-id error.\nbefore:\n%s\nafter:\n%s",
			req.Snapshot, after)
	}
}
```
