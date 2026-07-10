# Scenario

**Feature**: --limit N passed through to API and reflected in output size

```
slack-msg history --limit 2 -> API returns newest 2 -> CLI prints those 2 oldest→newest
```

## Preconditions

- slacktest history handler honors `limit` form field.

## Steps

1. Attach default slacktest.
2. Leaf sets `--limit 2`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	apiURL, err := ensureSlackTestServer(t)
	if err != nil {
		return err
	}
	req.SlackAPIURL = apiURL
	return nil
}
```
