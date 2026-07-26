---
label: e2e
---

## Expected

- Exit code 0.
- Stdout contains `DISCOVERY_MATCH_B` from the session whose first user chunk is `run ls`.
- Stdout does **not** contain `DISCOVERY_MATCH_A` (wrong-prompt session must not be tailed).

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	if !resp.StreamProbeSeen || !strings.Contains(resp.Stdout, discoveryMarkerB) {
		t.Fatalf("expected streamed marker from matching session %q; stdout:\n%s\nstderr:\n%s",
			discoveryMarkerB, resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stdout, discoveryMarkerA) {
		t.Fatalf("wrong session tailed: stdout contains %q:\n%s", discoveryMarkerA, resp.Stdout)
	}
}
```