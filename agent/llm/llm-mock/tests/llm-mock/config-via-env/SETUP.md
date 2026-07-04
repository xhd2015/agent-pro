# Scenario

**Feature**: LLM_MOCK_CONFIG env loading

```
server <- LLM_MOCK_CONFIG -> matched response
```

## Steps
1. Write config JSON to a temp file.
2. Do NOT pass `--config`; instead set `LLM_MOCK_CONFIG` env var pointing to the config file.
3. Send a matching request.
4. The server loads config from the env var and returns the configured response.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "env config test",
        "index": -1
      },
      "response": {
        "content": "env config works",
        "finish_reason": "stop"
      }
    }
  ]
}`
    req.UseEnvConfig = true
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"user","content":"env config test"}]}`,
    }
    return nil
}
```
