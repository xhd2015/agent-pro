---
label: unit
explanation: daemon lock probe with two instances; message uses slack-msg
---

## Expected

- First instance stays running until SIGTERM.
- Second instance exits non-zero.
- Second stderr contains `another slack-msg is already running`.

## Exit Code

0 (first instance stopped cleanly after probe)

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
	if resp.SecondExitCode == 0 {
		t.Fatalf("expected second instance non-zero exit; stdout=%q stderr=%q", resp.SecondStdout, resp.SecondStderr)
	}
	combined := resp.SecondStdout + resp.SecondStderr
	if !strings.Contains(combined, "another slack-msg is already running") {
		t.Fatalf("second instance output missing singleton message:\n%s", combined)
	}
}
```
