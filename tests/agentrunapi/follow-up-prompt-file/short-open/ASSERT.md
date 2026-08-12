## Expected

- No API error.
- Follow-up line contains open profile + separator `--` and prompt `hello`.
- Follow-up does **not** contain `--prompt-file`.
- `PromptSpillDir` has no auto-spill files.
- No `--new-terminal`.

## Side Effects

- None required (pure argv for short prompts).

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
	assertOpenProfile(t, fu, "sess-short-open")
	assertNoPromptFileFlag(t, fu)
	assertContains(t, fu, "--")
	assertContains(t, fu, fixtureShortPrompt)
	// Prompt is after `--`, not only as a substring of some other token.
	body := promptBodyAfterDashDash(fu)
	if !strings.Contains(body, fixtureShortPrompt) {
		t.Fatalf("short prompt must appear after `--`; body=%q line=%q", body, fu)
	}
	assertSpillDirEmpty(t, resp)
}
```
