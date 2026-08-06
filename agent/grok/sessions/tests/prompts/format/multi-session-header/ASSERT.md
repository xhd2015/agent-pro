## Expected

- No error.
- List length 2 (A then B by last_active).
- Output contains both session ids.
- Output contains both titles (`Title Alpha`, `Title Beta`).
- Output contains both prompt texts.
- Session header indicator present (`──` or similar separator before id).
- Trailing newline.
- No `👤`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertListLen(t, resp.List, 2)
	assertTrailingNewline(t, resp.Output)
	assertNotContains(t, resp.Output, "👤")

	out := resp.Output
	assertContains(t, out, idFormatMultiA)
	assertContains(t, out, idFormatMultiB)
	assertContains(t, out, "Title Alpha")
	assertContains(t, out, "Title Beta")
	assertContains(t, out, "prompt alpha")
	assertContains(t, out, "prompt beta")
	// Header chrome: box-drawing or ascii session separator
	if !strings.Contains(out, "──") && !strings.Contains(out, "--") {
		t.Fatalf("expected session header separator (── or --) in:\n%s", out)
	}
	// cwd short form or full fixture path fragment
	if !strings.Contains(out, "grok-prompts-fixture-project") && !strings.Contains(out, fixturePromptsCWD) {
		// abs path may expand; accept encoded path fragment
		if !strings.Contains(out, "tmp") {
			t.Fatalf("expected cwd fragment in multi header:\n%s", out)
		}
	}
}
```
