# Scenario

**Feature**: default global target writes to $HOME/.config/opencode

```
# no --dir: global target, anthropic shape, name defaults to id, one model
agent-pro opencode config add-provider --id myprov --base-url https://api.example.com/v1 --api-shape anthropic --model sonnet
doctest <- $HOME/.config/opencode/opencode.json : provider.myprov = { npm:@ai-sdk/anthropic, name:myprov, options.baseURL, models.sonnet }
```

## Preconditions

- `--id myprov`, `--base-url`, `--api-shape anthropic`, one `--model sonnet`.
- No `--dir`, no `--name`.

## Steps

1. Set `req.Args` to the full add-provider command.
2. Run with isolated `HOME`.
3. Assert the config file landed at `$HOME/.config/opencode/opencode.json`.

## Context

- Merges: global-default target, anthropic→`@ai-sdk/anthropic`, name defaults
  to id, and a single model entry.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"opencode", "config", "add-provider",
		"--id", "myprov",
		"--base-url", "https://api.example.com/v1",
		"--api-shape", "anthropic",
		"--model", "sonnet",
	}
	return nil
}
```
