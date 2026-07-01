# Scenario

**Feature**: duplicate provider id is rejected and the file is left unchanged

```
# provider[dup] already exists -> error mentions dup, file byte-for-byte unchanged
pre-config: { "provider": { "dup": {...} } }
agent-pro opencode config add-provider --id dup ... -> error -> exit != 0
doctest <- file bytes identical before/after
```

## Preconditions

- A config file pre-exists at the global target with `provider.dup` already
  present (a complete entry).
- The command is run with `--id dup` (colliding with the pre-existing id).

## Steps

1. Compute the resolved global config path and write the pre-existing config
   (with `provider.dup`) directly in Setup.
2. Snapshot the file bytes into `req.Snapshot`.
3. Set `req.Args` with `--id dup` (do NOT set `req.PreConfig`, since Setup
   wrote the file itself).
4. Run and assert non-zero exit + stderr mentions `dup`.
5. Assert the file bytes equal `req.Snapshot` (unchanged).

## Context

- Unlike the other error leaves, the file already exists here, so the
  assertion is byte-for-byte equality (proving no temp+rename landed), not
  absence.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"opencode", "config", "add-provider",
		"--id", "dup",
		"--base-url", "https://api.example.com/v1",
		"--api-shape", "anthropic",
		"--model", "m1",
	}
	// Write the pre-existing config (with provider.dup) directly so we can
	// snapshot exact bytes. Run will not touch PreConfig (it is unset).
	configPath := resolveConfigPath(req.Home, req.Args)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	content := []byte(`{
  "provider": {
    "dup": {
      "npm": "@ai-sdk/anthropic",
      "name": "dup",
      "options": { "baseURL": "https://old.example.com/v1" },
      "models": { "old": { "name": "old" } }
    }
  }
}
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		return err
	}
	req.Snapshot = content
	return nil
}
```
