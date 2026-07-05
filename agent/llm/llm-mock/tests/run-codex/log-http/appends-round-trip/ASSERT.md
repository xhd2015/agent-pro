## Expected

- Exit code 0.
- Log file at `LogHTTPPath` exists with at least 1 JSONL line.
- Each line is an HTTP exchange record with `request` and `response` objects.
- At least one line has `request.path` containing `/v1/responses`.
- That line has `response.status` = `200`.

## Side Effects

- Distinct from `--log-events` AgentEvent JSONL (`type` field).

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.LogHTTPLines) < 1 {
		t.Fatalf("log-http file: want >=1 JSONL line, got %d\ncontent:\n%s",
			len(resp.LogHTTPLines), resp.LogHTTPContent)
	}

	records, parseErr := parseHTTPExchangeMaps(resp.LogHTTPLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	var found bool
	for i, rec := range records {
		reqObj := rec["request"].(map[string]any)
		path, _ := reqObj["path"].(string)
		if !strings.Contains(path, "/v1/responses") {
			continue
		}
		respObj := rec["response"].(map[string]any)
		status, _ := respObj["status"].(float64)
		if int(status) != 200 {
			t.Fatalf("line %d: response.status = %v, want 200", i+1, respObj["status"])
		}
		found = true
		break
	}
	if !found {
		t.Fatalf("no log line with request.path containing /v1/responses and status 200 in:\n%s",
			resp.LogHTTPContent)
	}
}
```