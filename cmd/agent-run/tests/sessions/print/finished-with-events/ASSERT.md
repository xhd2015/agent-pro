## Expected Output

```
<contains>
web_test123
Events:
Hello from test
Done (session finished)
</contains>
```

## Expected

- Exit code 0.
- Stdout is formatted trace text (session header, numbered event lines, message bodies).
- Stdout is not raw NDJSON-only output.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assert.Output(t, resp.Stdout, `
<contains>
web_test123
Events:
Hello from test
Second trace line
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