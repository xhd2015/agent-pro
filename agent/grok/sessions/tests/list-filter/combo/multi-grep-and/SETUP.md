# Scenario

**Feature**: ListWithOptions multi-grep presence filter is same-unit AND

```
# idKeep title has both tokens; idSplit has one token in title and other only in
# a different field would still fail if not same unit — here: idPartial title
# has only one token; idKeep has both in title
GrepSet + Grep=[MG_A, MG_B] -> only idKeep
```

## Preconditions

- Grep presence uses the same search family as ListWithGrep (title field AND).
- No place/recent filters — isolate multi-grep.

## Steps

1. Write idKeep with both tokens in title.
2. Write idPartial with only MG_A in title.
3. GrepSet with both patterns.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.GrepSet = true
	req.Grep = []string{"MG_A", "MG_B"}
	writeListSession(t, req.GrokHome, idA1, atFixed(-20*time.Minute), cwdA, "has MG_A and MG_B together")
	writeListSession(t, req.GrokHome, idA2, atFixed(-10*time.Minute), cwdA, "has MG_A only")
	writeListSession(t, req.GrokHome, idB1, atFixed(-5*time.Minute), cwdB, "has MG_B only")
	return nil
}
```
