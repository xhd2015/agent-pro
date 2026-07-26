# Scenario

**Feature**: index=-1 sequential match

```
POST -> sequential counter -> Hello, world!
```

## Steps
1. Write config with a single exchange: index=-1, role="user", content="hello".
2. Send a chat request matching role and content.
3. The server matches sequentially and returns the configured response.

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
        "content": "Hello, world!",
        "finish_reason": "stop"
      }
    }
  ]
}`
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}`,
    }
    return nil
}
```
