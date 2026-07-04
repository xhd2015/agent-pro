## Expected

- PTY capture contains `RUN_OK` from the child shell command.
- Host does not need to print session id on attach (separate leaf covers silence).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	combined := resp.Combined
	if combined == "" {
		combined = resp.Stdout
	}
	if !strings.Contains(combined, "RUN_OK") {
		t.Fatalf("PTY output missing RUN_OK, got %q", combined)
	}
	assert.Output(t, combined, `---
version: 2
---
...1 lines omitted...
RUN_OK`)
}
```