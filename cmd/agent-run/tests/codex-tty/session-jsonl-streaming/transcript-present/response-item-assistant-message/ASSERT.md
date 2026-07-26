---
label: e2e
---

## Expected Output

```text
<contains>
JSONL_RESPONSE_ITEM_ASSISTANT
</contains>
```

## Expected

- Exit code 0.
- Stdout contains text from assistant `response_item.message` content items.

## Side Effects

- The assistant message is persisted in the codex-tty session events.

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
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assert.Output(t, resp.Stdout, `` +
`<contains>
JSONL_RESPONSE_ITEM_ASSISTANT
</contains>`)
	_, lines := findCodexTTYEventsJSONL(t, req.Home)
	if !eventsContainSubstring(t, lines, codexResponseItemText) {
		t.Fatalf("expected response_item assistant text in events.jsonl:\n%s", strings.Join(lines, "\n"))
	}
}
```
