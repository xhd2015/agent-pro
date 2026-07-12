## Expected

- Exit code 0.
- `KOOL_ITERM2_SCRIPT_OUT` AppleScript written with ModeForceNew (`create window`, no `create tab`).
- Follow-up command in script contains agent-run (or path), `run`, `--auto-send-or-resume`,
  session-id, and the prompt after a `--` separator; does **not** contain `--new-terminal`.
- No in-process provider spawn (argv probe absent/empty).

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)

	script := readItermScript(t, req)
	assertItermForceNewScript(t, script)

	// Follow-up command tokens (shell-quoted; match substrings).
	assertContains(t, script, "run")
	assertContains(t, script, "--auto-send-or-resume")
	assertContains(t, script, req.SessionID)
	assertContains(t, script, "--session-id")
	assertContains(t, script, req.FollowupPrompt)
	if strings.Contains(script, "--new-terminal") {
		t.Fatalf("follow-up must strip --new-terminal; script:\n%s", script)
	}
	// Prefer agent-run binary path or name in follow-up.
	assertContainsAny(t, script, "agent-run", req.AgentRun)

	// Non-empty prompt must use `--` separator before prompt text.
	hasSep := false
	for _, line := range strings.Split(script, "\n") {
		if !strings.Contains(line, "write text") {
			continue
		}
		if !strings.Contains(line, req.FollowupPrompt) {
			continue
		}
		// Separator forms: " -- ", `"--"`, `'--'` before prompt.
		promptIdx := strings.Index(line, req.FollowupPrompt)
		for _, sep := range []string{` -- `, `"--"`, `'--'`, `"-- `} {
			if i := strings.Index(line, sep); i >= 0 && i < promptIdx {
				hasSep = true
				break
			}
		}
	}
	if !hasSep {
		// Fallback: whole script has `--` token near prompt (shell-quoted command line).
		if strings.Contains(script, " -- "+req.FollowupPrompt) ||
			strings.Contains(script, `"--" `) ||
			strings.Contains(script, `'--' `) {
			hasSep = true
		}
	}
	if !hasSep {
		t.Fatalf("follow-up must include `--` before prompt %q; script:\n%s", req.FollowupPrompt, script)
	}

	assertNoInProcessProviderSpawn(t, req, resp)

	// Soft: launcher typically does not create meta (child in iTerm does).
	metaPath := metaJSONPath(req.Home, defaultRunner, req.SessionID)
	if fileExists(metaPath) {
		t.Logf("meta.json present at launcher exit (acceptable if implementer creates early): %s", metaPath)
	}
}
```
