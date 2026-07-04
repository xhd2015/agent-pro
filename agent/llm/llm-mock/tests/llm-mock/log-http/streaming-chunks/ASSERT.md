## Expected

- HTTP 200 SSE response on the chat completion request.
- `LogHTTPFile` exists with exactly 1 JSONL line.
- Line has `request.path` = `"/v1/chat/completions"`.
- Line has `response.stream` = `true` and `response.status` = `200`.
- `response.chunks` is a non-empty string array.
- At least one chunk contains the `data:` SSE prefix.
- At least one chunk contains `data: [DONE]`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	r := resp.Responses[0]
	if r.StatusCode != 200 {
		t.Fatalf("expected HTTP 200, got %d", r.StatusCode)
	}

	if len(resp.LogHTTPLines) != 1 {
		t.Fatalf("log-http: want 1 JSONL line, got %d\ncontent:\n%s",
			len(resp.LogHTTPLines), resp.LogHTTPContent)
	}

	records, parseErr := parseHTTPExchangeMaps(resp.LogHTTPLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	rec := records[0]

	reqObj := rec["request"].(map[string]any)
	if reqObj["path"] != "/v1/chat/completions" {
		t.Fatalf("request.path = %v, want /v1/chat/completions", reqObj["path"])
	}

	respObj := rec["response"].(map[string]any)
	if stream, _ := respObj["stream"].(bool); !stream {
		t.Fatalf("response.stream = false, want true")
	}
	status, _ := respObj["status"].(float64)
	if int(status) != 200 {
		t.Fatalf("response.status = %v, want 200", respObj["status"])
	}

	chunksRaw, ok := respObj["chunks"].([]any)
	if !ok || len(chunksRaw) == 0 {
		t.Fatalf("response.chunks must be non-empty array, got %#v", respObj["chunks"])
	}

	var hasDataPrefix bool
	var hasDone bool
	for _, c := range chunksRaw {
		chunk, _ := c.(string)
		if strings.Contains(chunk, "data:") {
			hasDataPrefix = true
		}
		if strings.Contains(chunk, "data: [DONE]") {
			hasDone = true
		}
	}
	if !hasDataPrefix {
		t.Fatalf("no chunk contains data: prefix in %#v", respObj["chunks"])
	}
	if !hasDone {
		t.Fatalf("no chunk contains data: [DONE] in %#v", respObj["chunks"])
	}
}
```