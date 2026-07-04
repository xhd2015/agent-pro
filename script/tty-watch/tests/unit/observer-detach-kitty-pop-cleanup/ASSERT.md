## Expected

- `observerTTYDetachCleanup` must pop grok's kitty keyboard protocol push (`\x1b[?u`)
  with `\x1b[<u`, not rely on `\x1b[?0u` alone.
- Without the pop, iTerm2 keeps delivering kitty key events after detach and typing
  shows fragments like `d0;1:3u` and `a7;1:3u`.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.SourceCheckOK {
		t.Fatalf("could not analyze attach.go: %s", resp.SourceCheckNote)
	}
	if !resp.KittyPopCleanupInSrc {
		t.Fatalf("attach.go observerTTYDetachCleanup missing \\x1b[<u kitty keyboard pop; iTerm2 keeps kitty mode after detach (%s)", resp.SourceCheckNote)
	}
}
```