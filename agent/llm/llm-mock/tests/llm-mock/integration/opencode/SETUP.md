## Steps
1. Create a temp working directory for opencode.
2. Write a mock config with an exchange matching "capital of France" → "Paris".
3. Set BinaryCmd to run `opencode run "What is the capital of France?" --model openai/gpt-4 --dir <temp>`.
4. Set BinaryEnv with OPENCODE_CONFIG_CONTENT (inline provider config pointing at the mock) and OPENCODE_DISABLE_PROJECT_CONFIG=true.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    workDir := t.TempDir()

    // Mock server config: one exchange matching "capital of France" → "Paris"
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "content": "capital of France",
        "index": -1
      },
      "response": {
        "content": "The capital of France is Paris.",
        "finish_reason": "stop"
      }
    }
  ]
}`

    // Inline opencode provider config pointing at the mock server.
    // __MOCK_PORT__ is replaced with the actual port by runBinary.
    opencodeConfig := `{
  "provider": {
    "openai": {
      "name": "openai",
      "npm": "@ai-sdk/openai",
      "options": {
        "baseURL": "http://localhost:__MOCK_PORT__/v1",
        "apiKey": "sk-test"
      },
      "models": {
        "gpt-4": {
          "id": "gpt-4",
          "name": "gpt-4",
          "tool_call": true
        }
      }
    }
  },
  "permission": {
    "question": "allow",
    "plan_enter": "allow",
    "plan_exit": "allow"
  }
}`

    req.BinaryCmd = []string{
        "opencode",
        "run",
        "What is the capital of France?",
        "--model", "openai/gpt-4",
        "--dir", workDir,
    }
    req.BinaryEnv = map[string]string{
        "OPENCODE_CONFIG_CONTENT":       opencodeConfig,
        "OPENCODE_DISABLE_PROJECT_CONFIG": "true",
    }
    return nil
}
```
