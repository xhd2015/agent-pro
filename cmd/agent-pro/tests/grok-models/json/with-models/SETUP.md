# Scenario

**Feature**: JSON Catalog includes sorted models and default

```
# same fixtures as human/with-models
List -> Catalog -> FormatJSON
```

## Preconditions

- Config + cache fixtures yield three model ids and default `grok-4.5`.

## Steps

1. Write config.toml and models_cache.json fixtures.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	writeGrokConfig(t, req.GrokHome, `
[models]
default = "grok-4.5"

[model."grok-4.5"]
context_window = 300000

[model."ais-glm-5-2"]
name = "AIS - GLM-5.2"
`)
	writeGrokCache(t, req.GrokHome, `{
  "models": {
    "grok-4.6": {},
    "grok-4.5": {}
  }
}`)
	return nil
}
```
