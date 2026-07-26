# Scenario

**Feature**: opencode with mock backend

```
opencode run -> mock -> Paris
```

## Steps
1. Create a temp working directory for opencode.
2. Isolate opencode from the user's real home: `HOME` and `OPENCODE_CONFIG_DIR` point at empty temp dirs so global plugins/hooks under `~/.config/opencode` are not loaded.
3. Write a mock config with an exchange matching "capital of France" → "Paris".
4. Set BinaryCmd to run `opencode run "What is the capital of France?" --model openai/gpt-4 --dir <temp>`.
5. Set BinaryEnv with OPENCODE_CONFIG_CONTENT (inline provider config pointing at the mock) and OPENCODE_DISABLE_PROJECT_CONFIG=true.

```go
import (
    "os"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    workDir := t.TempDir()
    isolatedHome := filepath.Join(workDir, "home")
    opencodeConfigDir := filepath.Join(workDir, "opencode-config")
    if err := os.MkdirAll(isolatedHome, 0755); err != nil {
        return err
    }
    if err := os.MkdirAll(opencodeConfigDir, 0755); err != nil {
        return err
    }

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
        "HOME":                            isolatedHome,
        "OPENCODE_CONFIG_DIR":             opencodeConfigDir,
        "OPENCODE_CONFIG_CONTENT":         opencodeConfig,
        "OPENCODE_DISABLE_PROJECT_CONFIG": "true",
    }
    return nil
}
```
