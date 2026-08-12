## Expected

- No API error (empty SessionID is not this case; SessionID is set).
- Follow-up has open profile.
- No `--prompt-file`.
- Spill dir empty.
- Current Open behavior preserved: typically still has `--` separator with
  empty prompt body (existing `BuildFollowUpCommand` open profile).

## Side Effects

- None.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	fu := resp.FollowUp
	assertOpenProfile(t, fu, "sess-empty-open")
	assertNoPromptFileFlag(t, fu)
	assertSpillDirEmpty(t, resp)
	// Existing Open profile always appends "--", prompt (prompt may be empty).
	assertContains(t, fu, "--")
}
```
