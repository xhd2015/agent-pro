## Expected

- Both GETs **200**.
- **all**: `counts.all=5`, `counts.running=1`, `counts.done=4`; `total=5`.
- **with-q**: `total=1` (q applied); **same** counts (`all=5`, `running=1`, `done=4`).

## Side Effects

- None (read-only).

## Errors

- Pre-impl: missing `counts` and/or `total` (RED). Counts collapsing with q is also RED.

```go
import (
	"testing"
)

func requireCounts(t *testing.T, body string, wantAll, wantRunning, wantDone int) {
	t.Helper()
	m := parseJSONMap(t, body)
	cm, ok := jsonCounts(m)
	if !ok {
		t.Fatalf("missing counts object: %q", truncate(body, 300))
	}
	all, okA := jsonFloat(cm, "all")
	run, okR := jsonFloat(cm, "running")
	done, okD := jsonFloat(cm, "done")
	if !okA || !okR || !okD {
		t.Fatalf("counts missing keys all/running/done: %#v", cm)
	}
	if int(all) != wantAll || int(run) != wantRunning || int(done) != wantDone {
		t.Fatalf("counts: got all=%v running=%v done=%v want all=%d running=%d done=%d",
			all, run, done, wantAll, wantRunning, wantDone)
	}
}

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}

	all := requireOK200(t, resp, "all")
	requireCounts(t, all.Body, 5, 1, 4)
	mAll := parseJSONMap(t, all.Body)
	totalAll, ok := jsonFloat(mAll, "total")
	if !ok {
		t.Fatal("all: missing total")
	}
	if int(totalAll) != 5 {
		t.Fatalf("all total: got %v want 5", totalAll)
	}

	withQ := requireOK200(t, resp, "with-q")
	requireCounts(t, withQ.Body, 5, 1, 4)
	mQ := parseJSONMap(t, withQ.Body)
	totalQ, ok := jsonFloat(mQ, "total")
	if !ok {
		t.Fatal("with-q: missing total")
	}
	if int(totalQ) != 1 {
		t.Fatalf("with-q total (must apply q): got %v want 1", totalQ)
	}
	sessions := sessionsFromBody(t, withQ.Body)
	if len(sessions) != 1 || sessionIDs(sessions)[0] != "sess-delta" {
		t.Fatalf("with-q sessions: got %v want [sess-delta]", sessionIDs(sessions))
	}
}
```
