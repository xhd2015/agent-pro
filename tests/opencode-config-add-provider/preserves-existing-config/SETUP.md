# Scenario

**Feature**: add-provider preserves pre-existing top-level keys and other providers

```
# pre-existing config has another provider + an unrelated top-level key
pre-config: { "provider": { "other": {...} }, "permission": {} }
agent-pro opencode config add-provider --id newprov --base-url ... --api-shape anthropic --model m1
doctest <- provider.other AND permission still present; provider.newprov added
```

## Preconditions

- A config file pre-exists at the global target containing:
  - `provider.other` (a complete provider entry), and
  - an unrelated top-level key `"permission": {}`.

## Steps

1. Set `req.PreConfig` to the pre-existing config JSON and let `Run` write it
   to the resolved global config path.
2. Set `req.Args` to add a new provider `newprov`.
3. Run and assert: `provider.newprov` is added, `provider.other` is unchanged,
   and the `permission` top-level key remains.

## Context

- This leaf verifies the command does a read-modify-write through
  `opencodecfg.ReadDir`/`Write` rather than overwriting the file.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.PreConfig = `{
  "provider": {
    "other": {
      "npm": "@ai-sdk/anthropic",
      "name": "other",
      "options": { "baseURL": "https://other.example.com/v1" },
      "models": { "o1": { "name": "o1" } }
    }
  },
  "permission": {}
}
`
	req.Args = []string{
		"opencode", "config", "add-provider",
		"--id", "newprov",
		"--base-url", "https://api.example.com/v1",
		"--api-shape", "anthropic",
		"--model", "m1",
	}
	return nil
}
```
