## Expected

- Exit code 0.
- Stdout contains scrollback-captured assistant text `hi`.
- `events.jsonl` contains the same assistant text.

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

	if !strings.Contains(resp.Stdout, "hi") {
		t.Fatalf("expected scrollback-captured hi on stdout:\n%s", resp.Stdout)
	}
	_, lines := findCodexTTYEventsJSONL(t, req.Home)
	if !eventsContainSubstring(t, lines, "hi") {
		t.Fatalf("expected scrollback-captured hi in events.jsonl:\n%s", strings.Join(lines, "\n"))
	}

	assert.Output(t, resp.Stdout, `` +
`<contains>
hi
</contains>`)
}
```
