# Scenario

**Feature**: JSON Catalog includes per-model reasoning and excludes hidden

```
# same fixtures as human/with-models
List -> Catalog -> FormatJSON
```

## Preconditions

- Config + cache fixtures with two list models and one hide model.

## Steps

1. Write config.toml and models_cache.json fixtures.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	writeCodexConfig(t, req.CodexHome, `model = "gpt-5.5"
`)
	writeCodexCache(t, req.CodexHome, `{
  "models": [
    {
      "slug": "gpt-5.6-sol",
      "display_name": "GPT-5.6-Sol",
      "default_reasoning_level": "medium",
      "visibility": "list",
      "supported_reasoning_levels": [
        {"effort": "low"},
        {"effort": "medium"},
        {"effort": "high"},
        {"effort": "xhigh"},
        {"effort": "max"},
        {"effort": "ultra"}
      ]
    },
    {
      "slug": "gpt-5.5",
      "display_name": "GPT-5.5",
      "default_reasoning_level": "xhigh",
      "visibility": "list",
      "supported_reasoning_levels": [
        {"effort": "low"},
        {"effort": "medium"},
        {"effort": "high"},
        {"effort": "xhigh"}
      ]
    },
    {
      "slug": "gpt-reserve",
      "display_name": "Reserve",
      "visibility": "hide",
      "supported_reasoning_levels": [{"effort": "medium"}]
    }
  ]
}`)
	return nil
}
```
