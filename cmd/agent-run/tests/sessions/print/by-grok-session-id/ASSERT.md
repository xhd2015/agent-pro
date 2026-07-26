---
label: e2e
---

## Expected Output

```
<contains>
print_gsid_s1
Events:
Hello from grok-session-id print
Done (session finished)
</contains>
```

## Expected

- Exit code 0.
- Stdout is formatted trace text (session header, event lines, message bodies).
- Stdout is not raw NDJSON-only output.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assert.Output(t, resp.Stdout, `
<contains>
print_gsid_s1
Events:
Hello from grok-session-id print
Second gsid print line
Done (session finished)
</contains>`)
	for _, line := range strings.Split(strings.TrimSpace(resp.Stdout), "\n") {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		if strings.HasPrefix(trim, "{") && strings.HasSuffix(trim, "}") {
			t.Fatalf("stdout should not be raw JSON lines; got line %q", trim)
		}
	}
}
```
