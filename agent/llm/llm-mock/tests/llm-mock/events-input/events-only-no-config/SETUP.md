# Scenario

**Feature**: events-only prefix — `LLM_MOCK_EVENTS_FILE` alone, no config file

```
# no --config / LLM_MOCK_CONFIG* — default empty exchanges
LLM_MOCK_EVENTS_FILE JSONL -> merged prefix script
POST -> response from events-only exchange
```

## Steps

1. Omit config file (`ConfigJSON` empty) and config env vars.
2. Set `LLM_MOCK_EVENTS_FILE` with one input exchange.
3. Send one matching chat completion request.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = ""
	req.EventsInputJSONL = `{"request":{"role":"user","content":"events-only-prompt","index":-1},"response":{"content":"from-events-only","finish_reason":"stop"}}` + "\n"
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"events-only-prompt"}]}`,
	}
	return nil
}
```