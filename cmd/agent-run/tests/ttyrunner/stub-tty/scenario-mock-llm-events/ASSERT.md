---
label: e2e
---

## Expected

- `events.jsonl` contains assistant event from scenario `llm_events`.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	lines := readEventsJSONL(t, resp.EventsFilePath)
	if len(lines) == 0 { t.Fatal("expected events.jsonl lines from llm_events") }
	found := false
	for _, ln := range lines {
		if strings.Contains(ln, "mock assistant reply") || strings.Contains(ln, "assistant") { found = true }
	}
	if !found { t.Fatalf("expected assistant event in %v", lines) }
}
```
