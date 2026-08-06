# Scenario

**Feature**: multi with --recent filters prompts by time window

```
# RecentSet=true
prompt included iff Timestamp in [Now-Recent, Now]
sessions with zero in-window prompts skipped (do not count toward Limit)
!LimitSet => no default session cap of 10
```

## Preconditions

- Parent sets RecentSet; leaf sets Recent duration and Limit flags.
- Window ends inclusive at Now and Now-Recent.

## Steps

1. Leaf seeds sessions with in/out of window prompt timestamps.
2. Assert filtering and limit interaction.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RecentSet = true
	return nil
}
```
