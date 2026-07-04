## Expected

- `List` with limit 20 returns exactly 20 sessions.
- Newest session id is `01900001-0000-7000-8000-000000000199`.
- The harness-measured `List` call (`resp.Elapsed`) completes within **500ms** —
  matching grok and opencode list latency when only 20 rows are requested.

## Errors

- None from List.

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 20 {
		t.Fatalf("len(sessions) = %d, want 20", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "01900001-0000-7000-8000-000000000199" {
		t.Fatalf("newest session id = %q, want 01900001-0000-7000-8000-000000000199", resp.Sessions[0].ID)
	}

	const budget = 500 * time.Millisecond
	if resp.Elapsed > budget {
		t.Fatalf("List took %v, want <= %v (currently parses every rollout file for TITLE/MSGS before applying limit)", resp.Elapsed, budget)
	}
}
```