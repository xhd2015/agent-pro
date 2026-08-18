# Scenario

**Feature**: cold Find rebuilds index and writes every observed UUID file

```
# no index/by-runner-session before call
seed grok-tty A and B (different UUIDs)
  -> FindByGrokSessionID(A)
  -> A.json and B.json exist; .gen exists
```

## Steps

1. Seed two grok-tty sessions with distinct provider UUIDs.
2. Op `find` for UUID A only (no prior warm).
3. Assert cache contains both UUID files and `.gen`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuidA = "11111111-1111-1111-1111-111111111111"
	const uuidB = "22222222-2222-2222-2222-222222222222"
	req.Op = "find"
	req.QueryID = uuidA
	req.Seeds = []SeedMeta{
		{SessionID: "cold-a", Runner: "grok-tty", RunnerSessionID: uuidA, Status: "finished"},
		{SessionID: "cold-b", Runner: "grok-tty", RunnerSessionID: uuidB, Status: "finished"},
	}
	return nil
}
```
