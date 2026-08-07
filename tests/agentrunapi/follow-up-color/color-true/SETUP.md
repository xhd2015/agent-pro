# Scenario

**Feature**: Color true emits `--color`

```
BuildFollowUpCommand(Color:true, Open, SessionID, Prompt)
  -> --color present as argv token
```

## Steps

1. Set `Color=true` with open profile defaults from root.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Color = true
	req.SessionID = "sess-color-true"
	req.Prompt = "color on follow-up"
	return nil
}
```
