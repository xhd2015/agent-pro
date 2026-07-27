## Expected Output

```text
<contains>
JSONL_DEDUPED_FINAL_MESSAGE
</contains>
```

## Expected

- Exit code 0.
- Stdout contains `JSONL_DEDUPED_FINAL_MESSAGE` exactly once.

## Side Effects

- Persisted events do not duplicate the assistant final answer.

## Errors

- No error is expected.

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
	assert.Output(t, resp.Stdout, `` +
`<contains>
JSONL_DEDUPED_FINAL_MESSAGE
</contains>`)
	if count := strings.Count(resp.Stdout, codexDedupedFinalText); count != 1 {
		t.Fatalf("expected %q once in stdout, got %d:\n%s", codexDedupedFinalText, count, resp.Stdout)
	}
	_, lines := findCodexTTYEventsJSONL(t, req.Home)
	if count := strings.Count(strings.Join(lines, "\n"), codexDedupedFinalText); count != 1 {
		t.Fatalf("expected %q once in events.jsonl, got %d:\n%s", codexDedupedFinalText, count, strings.Join(lines, "\n"))
	}
}
```
