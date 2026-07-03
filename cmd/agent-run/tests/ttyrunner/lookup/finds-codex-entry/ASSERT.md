## Expected

- Returns codex registry entry.
- Runner id is `codex-tty`.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.LookupRunnerID != "codex-tty" { t.Fatalf("runner: got %q want codex-tty", resp.LookupRunnerID) }
}
```
