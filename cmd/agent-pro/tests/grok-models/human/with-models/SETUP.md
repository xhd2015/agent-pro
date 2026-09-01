# Scenario

**Feature**: human listing marks configured default and unions config+cache

```
# config default + model keys, cache adds another id
List(home) -> Catalog{Default, Models sorted}

# FormatText marks default with "* "
FormatText -> "* grok-4.5" and "  other-id"
```

## Preconditions

- Config sets default `grok-4.5` and one extra model key.
- Cache adds `grok-4.6`.

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
