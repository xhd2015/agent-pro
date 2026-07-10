# Scenario

**Feature**: --thread uses conversations.replies

```
slack-msg history --thread TS -> conversations.replies -> chronological reply lines
```

## Preconditions

- slacktest replies handler returns `threadRepliesNewestFirst`.

## Steps

1. Attach default slacktest.
2. Leaf sets `--thread` with parent ts.

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
