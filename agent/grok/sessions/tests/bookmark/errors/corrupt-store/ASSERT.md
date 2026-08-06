## Expected

- List returns non-nil error (parse/corrupt).
- Store file still equals the corrupt marker (not replaced with a valid empty store).
- Subsequent `BookmarkGrok` also errors and does not clobber the corrupt body
  into a valid empty catalog.

## Errors

- corrupt/invalid JSON for list and pin

```go
import (
	"os"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertError(t, resp)

	raw := readStoreRaw(t, req.AgentProHome)
	if looksLikeEmptyValidStore(raw) {
		t.Fatalf("corrupt store was clobbered with valid empty store:\n%s", raw)
	}
	if !strings.Contains(raw, "not valid json") {
		t.Fatalf("store body lost corrupt marker after list error:\n%s", raw)
	}

	_, _, pinErr := sessions.BookmarkGrok(req.AgentProHome, req.GrokHome, req.SessionID, nil)
	if pinErr == nil {
		t.Fatal("BookmarkGrok on corrupt store expected error")
	}
	raw2, rerr := os.ReadFile(storePath(req.AgentProHome))
	if rerr != nil {
		t.Fatalf("read store after pin: %v", rerr)
	}
	body2 := string(raw2)
	if looksLikeEmptyValidStore(body2) {
		t.Fatalf("pin clobbered corrupt store into empty valid catalog:\n%s", body2)
	}
	if !strings.Contains(body2, "not valid json") {
		t.Fatalf("store body lost corrupt marker after pin error:\n%s", body2)
	}
}

func looksLikeEmptyValidStore(s string) bool {
	// Valid empty catalog would parse as versioned object with empty bookmarks
	// and would not contain our deliberate corrupt marker text.
	return strings.Contains(s, `"version"`) &&
		strings.Contains(s, `"bookmarks"`) &&
		!strings.Contains(s, "not valid json")
}
```
