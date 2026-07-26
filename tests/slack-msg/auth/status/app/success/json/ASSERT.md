---
label: unit
explanation: "app --json document with kind app, token_masked, note"
---

## Expected Output

```json
{"config":"(none)","kind":"app","ok":true,"token_masked":"xapp-...oken","note":"app-level token (Socket Mode / connections); not used for channels/send/history"}
```

## Expected

- Exit code 0.
- Single JSON document (trailing newline).
- Fields: `config`=`(none)`, `kind`=`app`, `ok`=true, `token_masked`, `note`.
- No raw full app token.
- Stderr empty.

## Exit Code

0

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, slackTestAppToken) {
		t.Fatalf("stdout must not contain raw app token %q:\n%s", slackTestAppToken, resp.Stdout)
	}
	if resp.Stdout == "" || resp.Stdout[len(resp.Stdout)-1] != '\n' {
		t.Fatalf("stdout must end with trailing newline, got %q", resp.Stdout)
	}
	var doc struct {
		Config      string `json:"config"`
		Kind        string `json:"kind"`
		Ok          bool   `json:"ok"`
		TokenMasked string `json:"token_masked"`
		Note        string `json:"note"`
	}
	if err := json.Unmarshal([]byte(resp.Stdout), &doc); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, resp.Stdout)
	}
	if doc.Config != "(none)" {
		t.Fatalf("config = %q, want (none)", doc.Config)
	}
	if doc.Kind != "app" || !doc.Ok {
		t.Fatalf("kind/ok = %q/%v, want app/true", doc.Kind, doc.Ok)
	}
	if doc.TokenMasked != slackTestAppTokenMasked {
		t.Fatalf("token_masked = %q, want %q", doc.TokenMasked, slackTestAppTokenMasked)
	}
	wantNote := "app-level token (Socket Mode / connections); not used for channels/send/history"
	if doc.Note != wantNote {
		t.Fatalf("note = %q, want %q", doc.Note, wantNote)
	}
}
```
