## Expected

- Returns grok registry entry.
- Runner id is `grok-tty`.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.LookupEntry == nil { t.Fatal("expected registry entry") }
	if resp.LookupRunnerID != "grok-tty" { t.Fatalf("runner: got %q want grok-tty", resp.LookupRunnerID) }
}
```
