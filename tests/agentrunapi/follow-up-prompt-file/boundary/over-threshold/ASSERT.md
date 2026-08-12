## Expected

- No API error.
- Follow-up contains `--prompt-file` with absolute path under `PromptSpillDir`.
- Spill content equals the 601×`a` prompt.
- Follow-up does not embed the long ASCII body.
- Open profile present; no `--new-terminal`.

## Side Effects

- Spill file under injected `PromptSpillDir`.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	fu := resp.FollowUp
	assertOpenProfile(t, fu, "sess-boundary-over")
	assertHasPromptFileFlag(t, fu)

	path := resp.PromptFilePath
	if path == "" {
		if p, ok := promptFileFlagValue(fu); ok {
			path = p
		}
	}
	assertAbsPath(t, path)
	rel, relErr := filepath.Rel(req.PromptSpillDir, path)
	if relErr != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("spill path %q must be under PromptSpillDir %q", path, req.PromptSpillDir)
	}

	wantBody := strings.TrimSpace(req.Prompt)
	assertEqual(t, "spill content", resp.SpillFileContent, wantBody)
	if strings.Contains(fu, wantBody) {
		t.Fatalf("follow-up must not embed 601-rune body; line len=%d", len(fu))
	}
	assertEqual(t, "rune count", runeCountTrimmed(req.Prompt), promptFileSpillMinRunes+1)
	if len(resp.SpillDirEntries) == 0 {
		t.Fatal("expected spill file under PromptSpillDir")
	}
}
```
