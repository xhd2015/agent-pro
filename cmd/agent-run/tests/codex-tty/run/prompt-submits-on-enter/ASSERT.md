## Expected

- Exit code 0.
- Captured stdout or events contain submitted prompt text `run ls` (not garbled CR artifacts).
- Output must **not** contain `UNSUBMITTED:run ls` (bare `\n` without `\r`).

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	combined := resp.Stdout + "\n" + resp.Stderr
	if strings.Contains(combined, "UNSUBMITTED:run ls") {
		t.Fatalf("prompt was injected with bare LF only (not submitted); output:\n%s", combined)
	}
	if !strings.Contains(resp.Stdout, "run ls") {
		_, lines := findCodexTTYEventsJSONL(t, req.Home)
		if !eventsContainSubstring(t, lines, "run ls") {
			t.Fatalf("expected submitted prompt run ls in stdout or events; stdout:\n%s\nevents:\n%s",
				resp.Stdout, strings.Join(lines, "\n"))
		}
	}
	if strings.Contains(combined, "un ls\nn ls") {
		t.Fatalf("prompt appears garbled (Enter not sent); output:\n%s", combined)
	}
}
```