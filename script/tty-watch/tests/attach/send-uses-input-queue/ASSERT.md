## Expected

- `send` exit code 0 (silent success).
- Attach handshake reports `attach_role` attacher while connected.
- PTY child capture contains exactly the injected marker bytes.
- No corruption, prefix, or suffix garbage in captured bytes.

```go
import (
	"bytes"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("send exit %d combined %q", resp.ExitCode, resp.Combined)
	}
	if !strings.Contains(resp.AttachOutput, `"attach_role":"attacher"`) {
		t.Fatalf("send+attach queue test requires attach_mode=attach attacher, handshake %q", resp.AttachOutput)
	}
	want := []byte(req.SendMessage)
	if len(want) == 0 {
		want = []byte("ATTACH_SEND_QUEUE_MARKER")
	}
	if !bytes.Equal(resp.InjectedBytes, want) {
		t.Fatalf("injected bytes %q (%v), want %q (%v)", resp.InjectedBytes, resp.InjectedBytes, want, want)
	}
	if bytes.Contains(resp.InjectedBytes, []byte{'\r'}) {
		t.Fatal("injected bytes must not contain unexpected \\r")
	}
	if bytes.Contains(resp.InjectedBytes, []byte{0x15}) {
		t.Fatal("injected bytes must not contain \\x15 Ctrl-U prefix")
	}
}
```