# Scenario

**Feature**: --name overrides the provider display name (default is id)

```
# --name given -> provider.name == display, NOT the id
agent-pro opencode config add-provider --id prov-a --name "My Display" --base-url https://api.example.com/v1 --api-shape anthropic --model m1
doctest <- provider.prov-a.name == "My Display"
```

## Preconditions

- `--name "My Display"` is provided and differs from `--id prov-a`.

## Steps

1. Set `req.Args` with `--name`.
2. Run and assert `provider.prov-a.name` == `My Display` (not `prov-a`).

## Context

- `global-default` covers the name-defaults-to-id behavior; this leaf covers
  the explicit `--name` override so the two paths are mutually exclusive.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"opencode", "config", "add-provider",
		"--id", "prov-a",
		"--name", "My Display",
		"--base-url", "https://api.example.com/v1",
		"--api-shape", "anthropic",
		"--model", "m1",
	}
	return nil
}
```
