## Expected

- JSON includes session_id plus title/cwd/kind; no bare-id line; stderr empty.

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	var got struct {
		SessionID string `json:"session_id"`
		Title     string `json:"title"`
		CWD       string `json:"cwd"`
		Kind      string `json:"kind"`
		LastActive string `json:"last_active"`
	}
	if err := json.Unmarshal([]byte(resp.Stdout), &got); err != nil {
		t.Fatalf("json decode: %v\n%s", err, resp.Stdout)
	}
	if got.SessionID != fixtureSessionID {
		t.Fatalf("session_id = %q, want %q", got.SessionID, fixtureSessionID)
	}
	if got.Title != fixtureDetailsTitle {
		t.Fatalf("title = %q, want %q", got.Title, fixtureDetailsTitle)
	}
	if got.CWD != fixtureDetailsCWD {
		t.Fatalf("cwd = %q, want %q", got.CWD, fixtureDetailsCWD)
	}
	if got.Kind != "sub" {
		t.Fatalf("kind = %q, want sub", got.Kind)
	}
	if strings.TrimSpace(got.LastActive) == "" {
		t.Fatal("last_active empty")
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr want empty, got %q", resp.Stderr)
	}
}
```
