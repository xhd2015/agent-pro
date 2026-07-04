## Expected

- Exit code 0.
- Stdout and stderr are empty on success.
- PTY child capture file contains exactly `follow up` (11 bytes).
- No `\r`, `\n`, or `\x15` (Ctrl-U) prefix/suffix in captured bytes.

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
	if resp.Stdout != "" || resp.Stderr != "" {
		t.Fatalf("success should be silent; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	want := []byte("follow up")
	if !bytes.Equal(resp.InjectedBytes, want) {
		t.Fatalf("injected bytes %q (%v), want %q (%v)", resp.InjectedBytes, resp.InjectedBytes, want, want)
	}
	if bytes.Contains(resp.InjectedBytes, []byte{'\r'}) {
		t.Fatal("injected bytes must not contain \\r suffix")
	}
	if bytes.Contains(resp.InjectedBytes, []byte{0x15}) {
		t.Fatal("injected bytes must not contain \\x15 Ctrl-U prefix")
	}
}
```