# Scenario

**Feature**: human listing marks configured default and unions config+cache

```
# config default + model keys, cache adds another id with display name
List(home) -> Catalog{Default, Models sorted with source/display_name}

# FormatText marks default with "* " and shows display names
FormatText -> "* grok-4.5  Grok 4.5"
```

## Preconditions

- Config sets default `grok-4.5` and one extra model key with `name`.
- Cache adds `grok-4.6` with `info.name` and fills `grok-4.5` display name.

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
    "grok-4.6": { "info": { "name": "Grok 4.6" } },
    "grok-4.5": { "info": { "name": "Grok 4.5" } }
  }
}`)
	return nil
}
```
