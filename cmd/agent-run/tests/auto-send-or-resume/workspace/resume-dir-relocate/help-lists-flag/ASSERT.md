## Expected

- Exit code 0 for `run -h`.
- `run -h` stdout contains exact flag `--allow-relocate-resume-session-dir`.
- `run -h` stdout ends with trailing newline `\n`.
- `resume -h` (second invocation) also documents the same flag; exit 0; trailing `\n`.

## Exit Code

0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, "--allow-relocate-resume-session-dir")
	assertTrailingNewline(t, resp.Stdout, "run help stdout")

	// Flag must also be on resume help (wired on resume subcommand).
	resumeResp, rErr := runAgentRun(t, req, "resume", "-h")
	if rErr != nil {
		t.Fatalf("resume -h: %v", rErr)
	}
	assertSuccess(t, resumeResp)
	assertContains(t, resumeResp.Stdout, "--allow-relocate-resume-session-dir")
	assertTrailingNewline(t, resumeResp.Stdout, "resume help stdout")
}
```
