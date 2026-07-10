---
label: unit
explanation: "multi-type --json soft-skips private missing_scope; public JSON + warning with see: topic"
---

## Expected Output

```json
{"channels":[{"id":"C0ALE44K5J6","name":"general","is_private":false,"is_member":true,"is_archived":false},{"id":"C0OTHERCHAN","name":"random","is_private":false,"is_member":false,"is_archived":false}]}
```

## Expected

- Exit code 0.
- JSON `channels` array: public only, sorted by name; archived omitted.
- Stderr contains
  `warning: skipped private channels (missing groups:read); see: slack-msg --help --topic add-missing-scope`
  (including with `--json`).
- Trailing newline on stdout (json.Encoder).

## Exit Code

0

```go
import (
	"encoding/json"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	assertStderrContains(t, resp, "warning: skipped private channels (missing groups:read); see: slack-msg --help --topic add-missing-scope")
	var doc struct {
		Channels []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			IsPrivate  bool   `json:"is_private"`
			IsMember   bool   `json:"is_member"`
			IsArchived bool   `json:"is_archived"`
		} `json:"channels"`
	}
	if err := json.Unmarshal([]byte(resp.Stdout), &doc); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, resp.Stdout)
	}
	want := []struct {
		id, name           string
		priv, member, arch bool
	}{
		{"C0ALE44K5J6", "general", false, true, false},
		{"C0OTHERCHAN", "random", false, false, false},
	}
	if len(doc.Channels) != len(want) {
		t.Fatalf("channels len = %d, want %d: %+v", len(doc.Channels), len(want), doc.Channels)
	}
	for i, w := range want {
		c := doc.Channels[i]
		if c.ID != w.id || c.Name != w.name || c.IsPrivate != w.priv || c.IsMember != w.member || c.IsArchived != w.arch {
			t.Fatalf("channels[%d] = {%q,%q,priv=%v,member=%v,arch=%v}, want {%q,%q,priv=%v,member=%v,arch=%v}",
				i, c.ID, c.Name, c.IsPrivate, c.IsMember, c.IsArchived, w.id, w.name, w.priv, w.member, w.arch)
		}
	}
	if resp.Stdout == "" || resp.Stdout[len(resp.Stdout)-1] != '\n' {
		t.Fatalf("stdout must end with trailing newline, got %q", resp.Stdout)
	}
}
```
