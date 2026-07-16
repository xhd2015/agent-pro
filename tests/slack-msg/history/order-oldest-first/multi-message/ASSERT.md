---
label: unit
explanation: human history lines oldest→newest despite newest-first API
---

## Expected Output

```
[1710000001.000100] U_OLDER: first message
[1710000002.000200] U_NEWER: second message
[1710000003.000300] U_NEWEST: third message
```

## Expected

- Exit code 0.
- Three human lines oldest→newest; trailing newline.
- Stderr empty.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
\[1710000001\.000100\] U_OLDER: first message
\[1710000002\.000200\] U_NEWER: second message
\[1710000003\.000300\] U_NEWEST: third message
`)
}
```
