---
label: e2e
---

## Expected

- Open stderr contains `grok-tty:` and `grok session`.
- Paris visible.

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
	if resp.Err != nil {
		t.Fatalf("open flow error: %v\n%s", resp.Err, resp.Open.Stderr)
	}
	c := resp.Open.Stderr + "\n" + resp.Open.Stdout
	if !strings.Contains(c, "grok-tty:") {
		t.Fatalf("stderr missing grok-tty: line:\n%s", c)
	}
	if !strings.Contains(c, "grok session") {
		t.Fatalf("stderr missing grok session line:\n%s", c)
	}
	if !resp.HasParis {
		t.Fatalf("want Paris; snap=%s", resp.ParisSnapshot)
	}
	if !resp.BoundOnOpen {
		t.Fatalf("expected BoundOnOpen; stderr=\n%s", resp.Open.Stderr)
	}
}
```
