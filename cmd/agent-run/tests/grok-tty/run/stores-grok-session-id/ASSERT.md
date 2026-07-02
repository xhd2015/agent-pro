## Expected

- Exit code 0.
- `AGENT_RUN_HOME/sessions/grok-tty/<id>/meta.json` exists.
- `runner_session_id` field equals the grok on-disk session UUID (`550e8400-e29b-41d4-a716-446655440000`).

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	path, meta := findGrokTTYMetaJSON(t, req.Home)
	if path == "" {
		t.Fatal("meta.json not found under sessions/grok-tty/")
	}
	got := resp.MetaRunnerSessionID
	if got == "" && meta != nil {
		if v, ok := meta["runner_session_id"].(string); ok {
			got = v
		}
	}
	if got != storeGrokUUID {
		t.Fatalf("runner_session_id = %q, want %q; meta path %s", got, storeGrokUUID, path)
	}
}
```