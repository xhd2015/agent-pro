## Expected

- Non-zero exit.
- Stderr contains `Error:` and `--set-config requires`.

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
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0\nstdout:\n%s", resp.Stdout)
	}
	assertContains(t, resp.Stderr, "Error:")
	assertContains(t, resp.Stderr, "--set-config requires")
	if strings.Contains(resp.Stderr, "FAKE_AGENT_INVOKED") {
		t.Fatalf("must not invoke agent:\n%s", resp.Stderr)
	}
}
```
