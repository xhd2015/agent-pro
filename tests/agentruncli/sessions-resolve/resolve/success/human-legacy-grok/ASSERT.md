## Expected

- Err nil.
- Stdout exactly `legacy-gsid\n`.

## Expected Output

```
legacy-gsid
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
legacy-gsid
`)
}
```
