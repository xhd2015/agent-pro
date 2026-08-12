---
label: e2e
---

## Expected

**Desired product behavior** (invert crime-scene “bug present” checks):

1. Host stream for production open attach mode (`open`) includes mouse CSI
   that real Grok (via llm-mock-run-grok) emitted.
2. Control: `attach_mode=attach` host stream still has mouse CSI (child enabled
   modes; proves fixture + scrollback path).
3. Host should also include alt-screen enter (`?1049h`).

```go
import (
	"bytes"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp == nil {
		if err != nil {
			t.Fatalf("run error (nil response): %v", err)
		}
		t.Fatal("nil response")
	}
	if err != nil {
		t.Fatalf("run error: %v\nstderr:\n%s", err, resp.Stderr)
	}

	if !HostHasMouseCSI(resp.ControlAttachBytes) {
		t.Fatalf("control attach_mode=attach missing mouse CSI (fixture broken: llm-mock-run-grok/real grok); bytes=%d preview=%q stderr_tail=%q",
			len(resp.ControlAttachBytes), previewBytes(resp.ControlAttachBytes, 200), tailStr(resp.Stderr, 400))
	}

	if !HostHasMouseCSI(resp.HostBytes) {
		t.Fatalf("desired: host attach_mode=%q with llm-mock-run-grok must preserve mouse CSI (?1000h/?1002h/?1006h); got bytes=%d esc=%d preview=%q control_mouse=true",
			req.HostAttachMode, len(resp.HostBytes), bytes.Count(resp.HostBytes, []byte{0x1b}), previewBytes(resp.HostBytes, 240))
	}

	if !bytes.Contains(resp.HostBytes, CsiAltScreen) {
		t.Fatalf("desired: host attach_mode=%q should enter alt-screen (?1049h); bytes=%d preview=%q",
			req.HostAttachMode, len(resp.HostBytes), previewBytes(resp.HostBytes, 240))
	}
}

func previewBytes(b []byte, n int) string {
	if len(b) == 0 {
		return ""
	}
	if n > len(b) {
		n = len(b)
	}
	var out []byte
	for _, c := range b[:n] {
		if c == 0x1b {
			out = append(out, []byte("<ESC>")...)
			continue
		}
		if c >= 32 && c < 127 || c == '\n' || c == '\r' || c == '\t' {
			out = append(out, c)
		} else {
			out = append(out, '.')
		}
	}
	return string(out)
}

func tailStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
```
