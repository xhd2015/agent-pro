# Scenario

**Feature**: random fallback first response must not block on synchronous probe execution

```
config exchanges[] empty -> first POST triggers generator fallback
server cwd = repo root (like llm-mock run grok in a worktree)
POST #1 must return HTTP 200 within 3s (not hang on grep/bash probes)
```

## Steps

1. Config JSON with `exchanges: []`.
2. Start mock server with `ServerDir = RepoRoot` (not an empty temp dir).
3. Send one chat completion request with prompt `"hello"`.
4. HTTP client timeout is 3 seconds.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = `{
  "port": 8080,
  "exchanges": []
}`
	req.ServerDir = req.RepoRoot
	req.HTTPTimeout = 3 * time.Second
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"hello"}]}`,
	}
	return nil
}
```