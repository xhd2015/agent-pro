# Scenario

**Feature**: invalid --api-shape value is rejected

```
# --api-shape gemini (not anthropic|openai) -> error mentions valid values
agent-pro opencode config add-provider --id p --base-url ... --api-shape gemini --model m1 -> error
```

## Preconditions

- `--api-shape gemini` (a value outside the allowed set).

## Steps

1. Set `req.Args` with `--api-shape gemini`.
2. Run and assert non-zero exit + stderr mentions `gemini` (the bad value) or
  the valid values (`anthropic`/`openai`); no config written.

## Context

- Distinct from `missing-api-shape`: here the flag is present but its value is
  not in `{anthropic, openai}`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"opencode", "config", "add-provider",
		"--id", "p",
		"--base-url", "https://api.example.com/v1",
		"--api-shape", "gemini",
		"--model", "m1",
	}
	return nil
}
```
