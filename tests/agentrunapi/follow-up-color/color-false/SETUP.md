# Scenario

**Feature**: Color false omits `--color`

```
BuildFollowUpCommand(Color:false, Open, SessionID, Prompt)
  -> no --color token
```

## Steps

1. Set `Color=false` with open profile.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Color = false
	req.SessionID = "sess-color-false"
	req.Prompt = "color off follow-up"
	return nil
}
```
