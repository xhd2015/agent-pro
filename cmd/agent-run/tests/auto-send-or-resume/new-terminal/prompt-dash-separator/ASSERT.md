## Expected

- Exit code 0.
- iTerm ForceNew script written.
- Follow-up command contains a `--` separator before the prompt text `-v explain`
  (or before `-v` / `explain`).
- Follow-up does **not** contain `--new-terminal`.

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
	if strings.Contains(script, "--new-terminal") {
		t.Fatalf("follow-up must strip --new-terminal; script:\n%s", script)
	}

	// Prompt text must appear in the follow-up command.
	assertContainsAny(t, script, "-v explain", "-v", "explain")

	// Require a standalone `--` argv token before the dash-leading prompt.
	// Accept common shell-quote shapes: `--` then prompt, or `-- '-v explain'`, etc.
	hasSep := false
	patterns := []string{
		`-- -v`,
		`-- '-v`,
		`-- "-v`,
		`"--" "-v`,
		`"--" '-v`,
		`'--' '-v`,
		`"--" -v`,
		`-- --`, // unlikely
	}
	for _, p := range patterns {
		if strings.Contains(script, p) {
			hasSep = true
			break
		}
	}
	// Fallback: find write text line(s) and require "--" token before "-v".
	if !hasSep {
		for _, line := range strings.Split(script, "\n") {
			if !strings.Contains(line, "write text") {
				continue
			}
			if !strings.Contains(line, "-v") {
				continue
			}
			idxSep := strings.Index(line, "--")
			idxV := strings.Index(line, "-v")
			// Need a `--` that is the separator (not part of --session-id etc.).
			// Search for " -- " or `"--"` before -v.
			if strings.Contains(line, " -- ") || strings.Contains(line, `"--"`) || strings.Contains(line, `'--'`) {
				// ensure that separator appears before -v in the line
				sepPos := strings.Index(line, " -- ")
				if sepPos < 0 {
					sepPos = strings.Index(line, `"--"`)
				}
				if sepPos < 0 {
					sepPos = strings.Index(line, `'--'`)
				}
				if sepPos >= 0 && idxV > sepPos {
					hasSep = true
					break
				}
			}
			_ = idxSep
		}
	}
	if !hasSep {
		t.Fatalf("follow-up must place `--` before dash-leading prompt; script:\n%s", script)
	}

	assertNoInProcessProviderSpawn(t, req, resp)
}
```
