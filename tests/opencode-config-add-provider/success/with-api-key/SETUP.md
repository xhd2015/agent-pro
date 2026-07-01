# Scenario

**Feature**: --api-key writes options.apiKey alongside options.baseURL

```
# --api-key given -> provider[id].options has both baseURL and apiKey
agent-pro opencode config add-provider --id keyprov --base-url https://api.example.com/v1 --api-shape anthropic --model m1 --api-key mysecret
doctest <- provider.keyprov.options == { "baseURL": "...", "apiKey": "mysecret" }
```

## Preconditions

- All mandatory flags supplied (`--id keyprov`, `--base-url`, `--api-shape anthropic`, one `--model`).
- Optional `--api-key mysecret` provided.

## Steps

1. Set `req.Args` to the full add-provider command including `--api-key mysecret`.
2. Run with isolated `HOME`.
3. Assert the shared success invariants via `assertSuccessCommon` (exit 0, id mentioned, provider entry present with `npm`/`options`/`models`).
4. Assert `provider.keyprov.options.apiKey == "mysecret"` AND `options.baseURL` is still the passed base-url.

## Context

- This leaf isolates the new optional `--api-key` factor. Sibling `global-default` covers the omitted-`--api-key` case (options has only `baseURL`); this leaf covers the present-`--api-key` case so the two paths are mutually exclusive.
- `assertSuccessCommon` is inherited from `success/SETUP.md` (this leaf is under `success/`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"opencode", "config", "add-provider",
		"--id", "keyprov",
		"--base-url", "https://api.example.com/v1",
		"--api-shape", "anthropic",
		"--model", "m1",
		"--api-key", "mysecret",
	}
	return nil
}
```
