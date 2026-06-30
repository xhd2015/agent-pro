## Expected

- Exit code 0.
- `events.jsonl` exists under `AGENT_RUN_HOME/sessions/<runner>/<session_id>/`.
- File lines match stdout NDJSON lines exactly (order preserved).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	stdoutLines := make([]string, 0)
	for _, line := range strings.Split(strings.TrimSpace(resp.Stdout), "\n") {
		if strings.TrimSpace(line) != "" {
			stdoutLines = append(stdoutLines, strings.TrimSpace(line))
		}
	}
	path, fileLines := findEventsJSONL(t, req.Home)
	if len(fileLines) != len(stdoutLines) {
		t.Fatalf("events.jsonl line count %d != stdout line count %d\nfile: %s", len(fileLines), len(stdoutLines), path)
	}
	for i := range stdoutLines {
		if fileLines[i] != stdoutLines[i] {
			t.Fatalf("line %d mismatch:\nstdout: %s\nfile:   %s", i, stdoutLines[i], fileLines[i])
		}
	}
}
```