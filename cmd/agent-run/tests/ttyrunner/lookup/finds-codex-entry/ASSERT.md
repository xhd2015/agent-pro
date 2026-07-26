## Expected

- Returns codex registry entry.
- Runner id is `codex-tty`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.LookupRunnerID != "codex-tty" { t.Fatalf("runner: got %q want codex-tty", resp.LookupRunnerID) }
}
```
