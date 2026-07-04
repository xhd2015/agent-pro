---
label: real-grok, slow
---

## Expected

- After real grok alt-screen modes, `\x1b[99;5u` detaches watch (some terminals encode Ctrl-C this way).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.GrokModesSeen {
		t.Fatalf("grok terminal modes not observed, output %q", resp.Combined)
	}
	if resp.TimedOut {
		t.Fatal("watch did not detach on \\x1b[99;5u within 3s after real grok alt-screen")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("watch exit code = %d, want 0", resp.ExitCode)
	}
}
```