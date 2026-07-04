# Scenario

**Feature**: Some terminals encode Ctrl-C as `\x1b[99;5u` after kitty protocol

```
# real grok alt-screen -> watch -> terminal sends letter-c+ctrl kitty sequence
harness -> run grok -> watch -> inject ESC[99;5u (not ESC[3;5u)
```

Gap test: documents that harness-injected `\x1b[3;5u` passes while alternate
kitty encodings (e.g. `\x1b[99;5u`) may still hang — closer to terminals whose
keyboard encoding differs from what we assumed.

```go
import (
	"os/exec"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("grok"); err != nil {
		t.Skip("grok not found in PATH")
	}
	req.Phase = "watch-ctrl-c-detaches-real-grok-99u"
	return nil
}
```