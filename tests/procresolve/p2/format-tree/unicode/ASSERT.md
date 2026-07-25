## Expected Output

```
200 /usr/local/bin/agent-run run --session-id=ignored-cli
├── 201 /usr/local/bin/agent-run serve --session-id=ignored-cli
│   └── 202 /usr/local/bin/grok
```

## Expected

- No error from Run.
- `TreeText` is a strict full match of the Unicode template above (trailing `\n`).
- Contains box-drawing connectors: `├──`, `└──`, and `│`.
- Contains pids `200`, `201`, `202` and cmd substrings `agent-run` / `grok`.

## Side Effects

- None (pure function).

## Errors

- None.

## Exit Code

N/A

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	if resp == nil {
		t.Fatal("response is nil")
	}
	// Strict line-by-line match (v3); trailing newline via closing backtick placement.
	assert.Output(t, resp.TreeText, `---
version: 3
---
200 /usr/local/bin/agent-run run --session-id=ignored-cli
├── 201 /usr/local/bin/agent-run serve --session-id=ignored-cli
│   └── 202 /usr/local/bin/grok
`)
	// Belt-and-suspenders: explicit connector presence for review readability.
	for _, glyph := range []string{"├──", "└──", "│"} {
		if !strings.Contains(resp.TreeText, glyph) {
			t.Fatalf("TreeText missing Unicode connector %q:\n%s", glyph, resp.TreeText)
		}
	}
}
```
