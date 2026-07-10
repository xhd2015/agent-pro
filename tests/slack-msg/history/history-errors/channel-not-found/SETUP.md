# Scenario

**Feature**: history unknown channel name

```
slack-msg history --channel #missing -> channel not found -> history failed:
```

## Steps

1. Default slacktest (no missing channel in list).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	apiURL, err := ensureSlackTestServer(t)
	if err != nil {
		return err
	}
	req.SlackAPIURL = apiURL
	req.Args = []string{
		"history",
		"--token", slackTestToken,
		"--channel", "#missing-channel",
	}
	return nil
}
```
