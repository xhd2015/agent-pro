# Scenario

**Feature**: EnrichInfo false leaves GrokTitle/GrokModel empty despite lookup

```
EnrichInfo=false + same grok hit + LookupGrokInfo available
  -> resolve still succeeds
  -> GrokTitle="", GrokModel="" (enrich gate off)
```

## Preconditions

- Parent installs hit + InjectLookup that would return title/model if called.
- This leaf sets EnrichInfo=false explicitly.

## Steps

1. Set `req.EnrichInfo = false`.
2. Assert hard hit succeeds and title/model are empty strings.

## Context

- Proves enrichment is opt-in; having LookupGrokInfo set must not auto-fill when
  EnrichInfo is false.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.EnrichInfo = false
	return nil
}
```
