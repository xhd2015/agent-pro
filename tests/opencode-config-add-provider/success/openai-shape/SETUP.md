# Scenario

**Feature**: --api-shape openai maps npm to @ai-sdk/openai-compatible

```
# openai shape -> npm == @ai-sdk/openai-compatible (distinct from anthropic)
agent-pro opencode config add-provider --id oai --base-url https://api.example.com/v1 --api-shape openai --model gpt-4o
doctest <- provider.oai.npm == @ai-sdk/openai-compatible
```

## Preconditions

- `--api-shape openai` (the non-anthropic valid shape).

## Steps

1. Set `req.Args` with `--api-shape openai`.
2. Run and assert `provider.oai.npm` == `@ai-sdk/openai-compatible`.

## Context

- The anthropic shape is covered by `global-default`; this leaf covers the
  other valid api-shape so the mapping table is fully exercised.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"opencode", "config", "add-provider",
		"--id", "oai",
		"--base-url", "https://api.example.com/v1",
		"--api-shape", "openai",
		"--model", "gpt-4o",
	}
	return nil
}
```
