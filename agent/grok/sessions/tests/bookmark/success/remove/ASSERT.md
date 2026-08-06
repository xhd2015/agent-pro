## Expected

- First remove: no error.
- Store no longer has (grok, session_id) — either missing store file, empty
  bookmarks array, or entry absent.
- Second `RemoveBookmark` with same args returns error containing `not found`.

## Errors

- Second remove: not found.

```go
import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)

	path := storePath(req.AgentProHome)
	if _, err := os.Stat(path); err == nil {
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			t.Fatalf("read store: %v", rerr)
		}
		var doc struct {
			Bookmarks []bookmarkEntry `json:"bookmarks"`
		}
		if uerr := json.Unmarshal(b, &doc); uerr != nil {
			t.Fatalf("unmarshal store after remove: %v", uerr)
		}
		if _, ok := findEntry(doc.Bookmarks, "grok", req.SessionID); ok {
			t.Fatalf("bookmark still present after remove: %+v", doc.Bookmarks)
		}
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat store: %v", err)
	}

	err2 := sessions.RemoveBookmark(req.AgentProHome, req.Runner, req.SessionID)
	if err2 == nil {
		t.Fatal("second RemoveBookmark expected error, got nil")
	}
	msg := err2.Error()
	if !strings.Contains(strings.ToLower(msg), "not found") {
		t.Fatalf("second remove error %q missing not found", msg)
	}
}
```
