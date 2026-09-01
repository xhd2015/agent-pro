# Scenario

**Feature**: human listing marks default, shows reasoning, hides non-list models

```
# cache has list + hide entries; config default matches a list slug
List -> Catalog with only visibility=list models

# FormatText marks default and includes reasoning=[...]
FormatText -> "* gpt-5.5 ..." and no "gpt-reserve"
```

## Preconditions

- Config `model = "gpt-5.5"`.
- Cache has two `list` models and one `hide` model.

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
