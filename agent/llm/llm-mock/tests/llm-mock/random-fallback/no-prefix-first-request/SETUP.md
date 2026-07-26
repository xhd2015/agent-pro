# Scenario

**Feature**: empty prefix exchanges; first HTTP request uses random generator → 200

```
config exchanges[] empty -> prefix exhausted immediately
POST #1 -> GenerateEvents -> ActionThink (or next event) -> HTTP 200
```

## Steps

1. Config JSON with `exchanges: []` (no prefix script).
2. Send one chat completion request.
3. Server must return HTTP 200 (not HTTP 400 `no_match`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = `{
  "port": 8080,
  "exchanges": []
}`
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"random-fallback-first-prompt"}]}`,
	}
	return nil
}
```