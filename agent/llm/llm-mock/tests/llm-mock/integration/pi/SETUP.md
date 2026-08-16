# Scenario

**Feature**: pi with mock backend

```
pi -> mock -> Paris
```

## Steps
1. Create a temp PI_CODING_AGENT_DIR and write `models.json` inside it with the openai provider config pointing at the mock server.
2. Write a mock config with an exchange matching "capital of France" → "Paris" (role-match-any: empty role so pi's message format always matches).
3. Set BinaryCmd to run `pi --provider openai --model gpt-4 -p "What is the capital of France?"`.
4. Set BinaryEnv with PI_CODING_AGENT_DIR.

```go
import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    if _, err := exec.LookPath("pi"); err != nil {
        t.Skipf("pi not on PATH: %v", err)
    }
    // Create temp config directory for pi
    configDir := t.TempDir()

    // Write models.json with the openai provider pointing at the mock.
    // __MOCK_PORT__ is replaced with the actual port by runBinary.
    modelsJSON := `{
  "providers": {
    "openai": {
      "baseUrl": "http://localhost:__MOCK_PORT__/v1",
      "api": "openai-completions",
      "apiKey": "sk-test"
    }
  }
}`
    if err := os.WriteFile(filepath.Join(configDir, "models.json"), []byte(modelsJSON), 0644); err != nil {
        return fmt.Errorf("write models.json: %w", err)
    }

    // Mock server config: one exchange matching "capital of France" → "Paris"
    // Empty role = match-any, so pi's messages always match regardless of role.
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "content": "capital of France",
        "index": -1
      },
      "response": {
        "content": "Paris is the capital of France.",
        "finish_reason": "stop"
      }
    }
  ]
}`

    req.BinaryCmd = []string{
        "pi",
        "--provider", "openai",
        "--model", "gpt-4",
        "-p", "What is the capital of France?",
    }
    req.BinaryEnv = map[string]string{
        "PI_CODING_AGENT_DIR": configDir,
    }
    return nil
}
```
