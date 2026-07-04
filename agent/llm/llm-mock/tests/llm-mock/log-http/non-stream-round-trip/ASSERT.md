## Expected

- HTTP 200 on the chat completion request.
- `LogHTTPFile` exists with exactly 1 JSONL line.
- Line has `request.method` = `"POST"` and `request.path` = `"/v1/chat/completions"`.
- Line has `response.stream` = `false` and `response.status` = `200`.
- `response.body` is a JSON object whose `choices[0].message.content` is `"Paris"`.

## Exit Code

0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

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
	if reqObj["method"] != "POST" {
		t.Fatalf("request.method = %v, want POST", reqObj["method"])
	}
	if reqObj["path"] != "/v1/chat/completions" {
		t.Fatalf("request.path = %v, want /v1/chat/completions", reqObj["path"])
	}

	respObj := rec["response"].(map[string]any)
	if stream, _ := respObj["stream"].(bool); stream {
		t.Fatalf("response.stream = true, want false")
	}
	status, _ := respObj["status"].(float64)
	if int(status) != 200 {
		t.Fatalf("response.status = %v, want 200", respObj["status"])
	}

	body, ok := respObj["body"].(map[string]any)
	if !ok {
		t.Fatalf("response.body must be object, got %T %#v", respObj["body"], respObj["body"])
	}
	choices, ok := body["choices"].([]any)
	if !ok || len(choices) < 1 {
		t.Fatalf("response.body.choices missing: %#v", body)
	}
	choice := choices[0].(map[string]any)
	message, ok := choice["message"].(map[string]any)
	if !ok {
		t.Fatalf("choice.message missing: %#v", choice)
	}
	if message["content"] != "Paris" {
		t.Fatalf("response.body message.content = %v, want Paris", message["content"])
	}
}
```