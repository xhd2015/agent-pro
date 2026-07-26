## Expected

- Returns grok registry entry.
- Runner id is `grok-tty`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.LookupEntry == nil { t.Fatal("expected registry entry") }
	if resp.LookupRunnerID != "grok-tty" { t.Fatalf("runner: got %q want grok-tty", resp.LookupRunnerID) }
}
```
