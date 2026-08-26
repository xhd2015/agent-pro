## Expected

- Exit 0.
- Help mentions `--set-config`, `--show-config`, `--no-config`.
- Failing fake agent not invoked.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	combined := resp.Stdout + resp.Stderr
	for _, flag := range []string{"--set-config", "--show-config", "--no-config", "--color", "--no-color"} {
		assertContains(t, combined, flag)
	}
	if strings.Contains(resp.Stderr, "FAKE_AGENT_INVOKED") {
		t.Fatalf("help must not invoke agent:\n%s", resp.Stderr)
	}
}
```
