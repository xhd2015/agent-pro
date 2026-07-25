# Scenario

**Feature**: EnrichInfo true fills GrokTitle and GrokModel from LookupGrokInfo

```
EnrichInfo=true + Kind=grok + SessionID set
  -> LookupGrokInfo(GrokHome, SessionID) -> title, model
Result.GrokTitle / Result.GrokModel set
```

## Preconditions

- Parent installs hit fixture + injectable lookup returning fixture title/model.
- This leaf only enables EnrichInfo.

## Steps

1. Set `req.EnrichInfo = true`.
2. Assert hard hit fields plus GrokTitle/GrokModel match inject values.

## Context

- Title/model must come from LookupGrokInfo, not from open-file path parsing.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.EnrichInfo = true
	return nil
}
```
