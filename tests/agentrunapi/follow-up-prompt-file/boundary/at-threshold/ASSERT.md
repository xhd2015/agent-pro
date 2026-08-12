## Expected

- No API error.
- Follow-up has open profile and prompt body after `--` (inline).
- No `--prompt-file`.
- Spill dir empty.
- Prompt rune count is exactly 600 (sealed by Setup).

## Side Effects

- None.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	fu := resp.FollowUp
	assertOpenProfile(t, fu, "sess-boundary-at")
	assertNoPromptFileFlag(t, fu)
	assertSpillDirEmpty(t, resp)
	// Inline: the 600×'a' body should appear after `--`.
	body := promptBodyAfterDashDash(fu)
	if body != req.Prompt && !strings.Contains(fu, req.Prompt) {
		t.Fatalf("at-threshold prompt must stay inline after `--`; body len=%d want %d; line len=%d",
			len(body), len(req.Prompt), len(fu))
	}
	assertEqual(t, "rune count", runeCountTrimmed(req.Prompt), promptFileSpillMinRunes)
}
```
