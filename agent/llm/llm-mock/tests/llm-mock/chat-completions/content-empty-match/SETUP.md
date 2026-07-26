# Scenario

**Feature**: empty content matches any

```
POST any content -> empty-content exchange
```

## Steps
1. Write config with an exchange where `request.content` is empty (should match any message).
2. Send a request with arbitrary content.
3. The exchange matches because empty content means match-any.

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
        "role": "",
        "content": "",
        "index": -1
      },
      "response": {
        "content": "catch-all response",
        "finish_reason": "stop"
      }
    }
  ]
}`
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"user","content":"anything at all"}]}`,
    }
    return nil
}
```
