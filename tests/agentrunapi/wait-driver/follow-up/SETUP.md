# Scenario

**Feature**: pure BuildFollowUpCommand (DriverBinary + child flags, no --new-terminal)

```
FollowUpOpts -> BuildFollowUpCommand -> shell-quoted line
```

## Steps

1. Set mode `follow_up`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "follow_up"
	return nil
}
```
