# Scenario

**Feature**: `MaxMessages` keeps the latest N messages

```
oldest → newest slice, MaxMessages=K
  -> last K messages only
  -> header showing all K of N when K==N; showing last K of N when K<N
```

## Preconditions

- Input order oldest → newest; trim drops from the **oldest** side.
- `MaxMessages=0` means no count cap (not exercised in this branch).
- Full multi uses `showing all K of N`; partial suffix uses `showing last K of N`.

## Steps

1. Branch Setup clears budget so only MaxMessages varies.
2. Leaf builds a multi-message list and sets MaxMessages.
3. Assert which ids/bodies appear and the K of N label.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// max-messages branch: no total-budget interaction; leaves set MaxMessages.
	req.Opts = msgfmt.Options{
		MaxPerMessageRunes: 0,
		MaxMessages:        0, // leaf overrides with a positive K
		TotalBudgetRunes:   0,
	}
	return nil
}
```
