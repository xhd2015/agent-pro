## Expected

- PTY capture contains `RUN_OK` from the child shell command.
- Host does not need to print session id on attach (separate leaf covers silence).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/script/tty-watch/ttywatchtest"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	raw := resp.Stdout
	if raw == "" {
		raw = resp.Combined
	}
	lines := ttywatchtest.VisibleContentLines(raw)
	for _, line := range lines {
		if line == "RUN_OK" {
			return
		}
	}
	t.Fatalf("PTY output missing RUN_OK line, lines=%v raw=%q", lines, raw)
}
```