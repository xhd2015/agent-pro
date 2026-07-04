# Scenario

**Feature**: `LLM_MOCK_EVENTS_FILE` input — server merges additional exchanges after config

```
mockconfig loader -> config exchanges + events JSONL
llm-mock HTTP server <- merged exchanges[]
```

## Preconditions

- Server mode accepts `LLM_MOCK_EVENTS_FILE` as **input** JSONL (not `--events-file` output).
- Events exchanges append after config `exchanges[]` (config file optional).

## Steps

1. Grouping `Setup` sets chat-completions endpoint defaults.
2. Leaf `Setup` provides config JSON (optional) + events input JSONL + HTTP requests.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Endpoint = "/v1/chat/completions"
    req.Method = "POST"
    return nil
}
```
