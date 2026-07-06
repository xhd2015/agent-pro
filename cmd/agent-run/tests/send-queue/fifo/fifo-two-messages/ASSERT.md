## Expected

- Blocking trigger send exits 0.
- Injection order is `fifo-message-A` then `fifo-message-B`.

## Exit Code

0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertSuccess(t, resp)
	want := []string{"fifo-message-A", "fifo-message-B"}
	if len(resp.InjectedMessages) < len(want) {
		t.Fatalf("expected injection order %v, got %v", want, resp.InjectedMessages)
	}
	for i, w := range want {
		if resp.InjectedMessages[i] != w {
			t.Fatalf("FIFO order mismatch at %d: want %v got %v", i, want, resp.InjectedMessages)
		}
	}
}
```