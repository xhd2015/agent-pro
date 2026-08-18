## Expected

- Err nil.
- Stdout is one JSON object ending with `\n` (Encoder newline OK).
- Required keys: `session_id`, `runner`, `runner_session_id`, `status` with
  values `hello-json`, `grok-tty`, the query UUID, `finished`.
- No ANSI; no pretty-indent required.

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.Err != nil {
		t.Fatalf("resolve --json error: %v", resp.Err)
	}
	out := resp.Stdout
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("JSON stdout must end with \\n, got %q", out)
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &obj); err != nil {
		t.Fatalf("stdout not JSON: %v\n%s", err, out)
	}
	wantUUID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa03"
	checks := map[string]string{
		"session_id":        "hello-json",
		"runner":            "grok-tty",
		"runner_session_id": wantUUID,
		"status":            "finished",
	}
	for k, want := range checks {
		got, ok := obj[k]
		if !ok {
			t.Fatalf("JSON missing key %q: %#v", k, obj)
		}
		s, _ := got.(string)
		if s != want {
			t.Fatalf("JSON %s=%v, want %q", k, got, want)
		}
	}
}
```
