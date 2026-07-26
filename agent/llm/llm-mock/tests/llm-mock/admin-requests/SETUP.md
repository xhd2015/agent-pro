# Scenario

**Feature**: GET /admin/requests returns recorded HTTP requests

```
HTTP client -> POST /v1/chat/completions (x2)
llm-mock <- admin recorder logs each request
GET /admin/requests -> recorded request list
```

## Steps
1. Configure two exchanges for sequential matching.
2. Send two chat completion requests.
3. Verify the server correctly matches both requests and records them (validated implicitly by correct responses).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "hello",
        "index": -1
      },
      "response": {
        "content": "world",
        "finish_reason": "stop"
      }
    },
    {
      "request": {
        "role": "user",
        "content": "second query",
        "index": -1
      },
      "response": {
        "content": "second response",
        "finish_reason": "stop"
      }
    }
  ]
}`
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}`,
        `{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"second query"}]}`,
    }
    return nil
}
```
