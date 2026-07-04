## Expected

- Exit code 0.
- Captured bytes are exactly `hello` with length 5.
- No trailing `\r` or `\n`.

```go
import (
	"bytes"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("send exit %d combined %q", resp.ExitCode, resp.Combined)
	}
	want := []byte("hello")
	if !bytes.Equal(resp.InjectedBytes, want) {
		t.Fatalf("injected bytes %q (%v), want %q (%v)", resp.InjectedBytes, resp.InjectedBytes, want, want)
	}
	if len(resp.InjectedBytes) != 5 {
		t.Fatalf("expected length 5, got %d bytes %q", len(resp.InjectedBytes), resp.InjectedBytes)
	}
	if bytes.HasSuffix(resp.InjectedBytes, []byte{'\r'}) || bytes.HasSuffix(resp.InjectedBytes, []byte{'\n'}) {
		t.Fatalf("must not append line ending suffix; got %q", resp.InjectedBytes)
	}
}
```