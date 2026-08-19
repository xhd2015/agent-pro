---
label: e2e
---

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
	"regexp"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	re := regexp.MustCompile(`(?m)^grok-tty:\s*session-\d+\s*$`)
	if !re.MatchString(resp.Stderr) {
		t.Fatalf("stderr missing grok-tty: session-N; stderr:\n%s", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "grok-tty:") {
		t.Fatalf("session id prefix must not appear on stdout:\n%s", resp.Stdout)
	}
}
```
