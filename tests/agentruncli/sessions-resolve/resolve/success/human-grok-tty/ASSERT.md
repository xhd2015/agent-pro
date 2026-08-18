## Expected

- Err nil.
- Stdout is exactly `hello-world` plus trailing newline (no runner prefix, no ANSI).

## Expected Output

```
hello-world
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.Err != nil {
		t.Fatalf("resolve error: %v", resp.Err)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
hello-world
`)
}
```
