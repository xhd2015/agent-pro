# Scenario

**Feature**: OpenAI Chat Completions encoder (`POST /v1/chat/completions`)

```
breakpoint slice -> chat.go encode -> choices[].message (content | tool_calls)
```

## Steps

1. Default `Endpoint` = `/v1/chat/completions`.
2. Leaves assert `tool_calls`, `finish_reason`, and merged think content.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Endpoint = "/v1/chat/completions"
	return nil
}
```