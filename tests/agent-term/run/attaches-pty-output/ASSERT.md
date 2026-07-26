## Expected

- Exit code 0.
- PTY capture contains `RUN_OK` before the final `session-N` line.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d combined %q", resp.ExitCode, resp.Combined)
	}
	assert.Output(t, resp.Combined, `
<contains>
RUN_OK
</contains>`)
	lines := strings.Split(strings.TrimSpace(resp.Combined), "\n")
	last := lines[len(lines)-1]
	if !strings.HasPrefix(last, "session-") {
		t.Fatalf("last line should be session id, got %q full %q", last, resp.Combined)
	}
}
```