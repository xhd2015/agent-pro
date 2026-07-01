# Scenario

**Feature**: no --model flag is rejected (at least one model required)

```
# --model omitted entirely -> error mentions model
agent-pro opencode config add-provider --id p --base-url ... --api-shape anthropic -> error
```

## Preconditions

- All mandatory flags except `--model` are supplied; no `--model` is given.

## Steps

1. Set `req.Args` with no `--model`.
2. Run and assert non-zero exit + stderr mentions `model`; no config written.

## Context

- `--model` is a repeatable slice with a minimum of one value; this leaf
  covers the zero-values case.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"opencode", "config", "add-provider",
		"--id", "p",
		"--base-url", "https://api.example.com/v1",
		"--api-shape", "anthropic",
	}
	return nil
}
```
