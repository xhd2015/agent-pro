## Expected

- Exit code 0.
- Stdout contains the fixture session id
  `019fabcdef-1234-5678-9abc-def012345678`.
- Stdout contains kind `grok` (as a JSON string value or field association).
- Stdout does **not** require Unicode tree connectors `├`, `└`, or `│`
  (JSON path must not depend on FormatTree glyphs).
- Stdout ends with a trailing newline.

## Side Effects

- None (fixture inject; no live process mutation).

## Errors

- None.

## Exit Code

- 0

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	assertStdoutEndsWithNewline(t, resp.Stdout)
	assertContains(t, resp.Stdout, fixtureGrokSessionID)
	assertContainsFold(t, resp.Stdout, "grok")

	// JSON body should parse as an object (or array); tolerate pretty-print.
	trim := strings.TrimSpace(resp.Stdout)
	if !json.Valid([]byte(trim)) {
		// Allow a single JSON value embedded with trailing whitespace only.
		t.Fatalf("stdout is not valid JSON:\n%s", resp.Stdout)
	}

	// Tree glyphs must not be required; if present, fail (JSON is not a tree dump).
	for _, glyph := range []string{"├", "└", "│"} {
		assertNotContains(t, resp.Stdout, glyph)
	}
}
```
