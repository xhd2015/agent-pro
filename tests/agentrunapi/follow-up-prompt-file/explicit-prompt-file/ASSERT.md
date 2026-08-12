## Expected

- No API error.
- Follow-up contains `--prompt-file` whose value equals the given absolute
  `PromptFile` path.
- `PromptSpillDir` remains empty (no auto-spill rewrite).
- Given file still has original body (not overwritten by empty Prompt).
- No prompt body after `--` from `req.Prompt` (empty / ignored).
- Open profile present; no `--new-terminal`.

## Side Effects

- None under `PromptSpillDir` (given file lives outside spill dir).

## Errors

- None.

## Exit Code

N/A

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	fu := resp.FollowUp
	assertOpenProfile(t, fu, "sess-explicit-pf")
	assertHasPromptFileFlag(t, fu)

	path := resp.PromptFilePath
	if path == "" {
		if p, ok := promptFileFlagValue(fu); ok {
			path = p
		}
	}
	assertEqual(t, "PromptFile path", path, req.PromptFile)
	assertAbsPath(t, path)

	// Must not auto-spill into PromptSpillDir.
	assertSpillDirEmpty(t, resp)

	// Given file must not be rewritten to empty when Prompt is empty.
	data, rerr := os.ReadFile(req.PromptFile)
	if rerr != nil {
		t.Fatalf("read given PromptFile: %v", rerr)
	}
	assertEqual(t, "given file body", string(data), "pre-spilled-body\n")

	// Empty / ignored Prompt should not appear as a required body after --.
	// Prefer file mode without trailing long/spurious prompt token.
	if body := promptBodyAfterDashDash(fu); body != "" && body != req.Prompt {
		// Allow empty body after `--` if implementer keeps Open's `--` separator.
		// Non-empty unexpected bodies fail.
		if req.Prompt == "" && body != "" {
			// empty Prompt: any leftover token after -- is suspicious for file mode
			// but a lone empty quoted string is ok; non-empty is not.
			t.Fatalf("file mode should not put a prompt body after `--`; body=%q line=%q", body, fu)
		}
	}
}
```
