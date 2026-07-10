# Scenario

**Feature**: --json emits structured chronological messages

```
slack-msg history --json -> single JSON document with messages oldest→newest
```

## Preconditions

- Same slacktest history fixture as human path.

## Steps

1. Attach default slacktest.
2. Leaf adds `--json`.

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
