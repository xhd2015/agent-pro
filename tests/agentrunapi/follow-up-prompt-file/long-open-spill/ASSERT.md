## Expected

- No API error.
- Follow-up contains `--prompt-file` (equals or two-token form).
- `--prompt-file` path is absolute and under `PromptSpillDir`.
- Spill file content equals `TrimSpace(Prompt)` (full long body).
- Follow-up does **not** embed the long prompt body after `--`.
- Line is much shorter than embedding the body (write-text safety).
- Open profile still present; no `--new-terminal`.

## Side Effects

- One (or more) spill file(s) under injected `PromptSpillDir`.

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
	assertOpenProfile(t, fu, "sess-long-spill")
	assertHasPromptFileFlag(t, fu)

	path := resp.PromptFilePath
	if path == "" {
		if p, ok := promptFileFlagValue(fu); ok {
			path = p
		}
	}
	assertAbsPath(t, path)
	// Prefer spill under injected PromptSpillDir.
	rel, relErr := filepath.Rel(req.PromptSpillDir, path)
	if relErr != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("spill path %q must be under PromptSpillDir %q", path, req.PromptSpillDir)
	}

	wantBody := strings.TrimSpace(req.Prompt)
	gotBody := resp.SpillFileContent
	if gotBody == "" {
		// Response may not have read yet; path already checked.
		t.Fatalf("spill file content empty or unreadable at %q", path)
	}
	assertEqual(t, "spill content", gotBody, wantBody)

	// Long body must not appear on the follow-up line (after -- or anywhere).
	if strings.Contains(fu, wantBody) {
		t.Fatalf("follow-up must not embed long prompt body; line len=%d body runes=%d",
			len(fu), runeCountTrimmed(wantBody))
	}
	bodyAfter := promptBodyAfterDashDash(fu)
	if strings.Contains(bodyAfter, wantBody) || bodyAfter == wantBody {
		t.Fatalf("long prompt must not appear after `--`; bodyAfter len=%d", len(bodyAfter))
	}
	// Line should be far shorter than embedding the multi-byte body.
	if len(fu) >= len(wantBody) {
		t.Fatalf("follow-up line len=%d should be << embedded body len=%d; line=%q",
			len(fu), len(wantBody), fu)
	}
	if len(resp.SpillDirEntries) == 0 {
		t.Fatal("expected at least one spill file under PromptSpillDir")
	}
}
```
