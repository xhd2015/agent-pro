---
label: unit
explanation: --json channels array sorted by name; archived excluded
---

## Expected Output

```json
{"channels":[{"id":"C0AGENTDBG1","name":"agent-pro-debug","is_private":true,"is_member":false,"is_archived":false},{"id":"C0ALE44K5J6","name":"general","is_private":false,"is_member":true,"is_archived":false},{"id":"C0OTHERCHAN","name":"random","is_private":false,"is_member":false,"is_archived":false}]}
```

## Expected

- Exit code 0.
- Single JSON document on stdout (trailing newline).
- `channels` sorted by name ascending; archived omitted.
- Fields: `id`, `name`, `is_private`, `is_member`, `is_archived`.
- Stderr empty.

## Exit Code

0

```go
import (
	"encoding/json"
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
		{"C0AGENTDBG1", "agent-pro-debug", true, false, false},
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
