# Scenario

**Feature**: repeatable --model produces a models map with every model id

```
# two --model flags -> models map has both keys, each {name: <id>}
agent-pro opencode config add-provider --id multiprov --base-url https://api.example.com/v1 --api-shape anthropic --model m1 --model m2
doctest <- provider.multiprov.models == { "m1": {"name":"m1"}, "m2": {"name":"m2"} }
```

## Preconditions

- Two `--model` flags: `m1` and `m2`.

## Steps

1. Set `req.Args` with two `--model` flags.
2. Run and assert `provider.multiprov.models` has exactly keys `m1` and `m2`,
   each value `{"name": "<id>"}`.

## Context

- `global-default` covers the single-model case; this leaf covers the
  repeatable-slice case so `flags.StringSlice("--model", ...)` is exercised
  with multiple values.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"opencode", "config", "add-provider",
		"--id", "multiprov",
		"--base-url", "https://api.example.com/v1",
		"--api-shape", "anthropic",
		"--model", "m1",
		"--model", "m2",
	}
	return nil
}
```
