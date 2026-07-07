## Expected

- CLI attach stdout contains `CLI_ATTACH_MARKER` after stdin is forwarded.
- Attach uses live registry-backed terminal (not snapshot-only scrollback).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.Stdout, "CLI_ATTACH_MARKER") {
		t.Fatalf("CLI attach did not relay stdin to PTY; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}
```
