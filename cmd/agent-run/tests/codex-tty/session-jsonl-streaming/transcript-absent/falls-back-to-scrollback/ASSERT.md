## Expected Output

```text
<contains>
SCROLLBACK_FALLBACK_WITHOUT_TRANSCRIPT
</contains>
```

## Expected

- Exit code 0.
- Stdout contains cleaned scrollback fallback text.
- Missing Codex rollout JSONL does not turn the run into an error.

## Side Effects

- The fallback assistant text is persisted in codex-tty events.

## Errors

- No error is expected.

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
	assert.Output(t, resp.Stdout, `
<contains>
SCROLLBACK_FALLBACK_WITHOUT_TRANSCRIPT
</contains>`)
	_, lines := findCodexTTYEventsJSONL(t, req.Home)
	if !eventsContainSubstring(t, lines, codexScrollbackFallbackText) {
		t.Fatalf("expected fallback text in events.jsonl:\n%s", strings.Join(lines, "\n"))
	}
}
```
