# Scenario

**Feature**: missing / zero target time formats as dash

```
FormatRelativeAge(now, time.Time{}) -> "-"
# empty updated_at/created_at after parse → zero Time → dash cell
```

## Preconditions

- Zero value `time.Time` is the sentinel for missing session timestamps.

## Steps

1. Set one case with zero target and want `-`.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Cases = []FormatCase{
		{Name: "zero_time", Target: time.Time{}, Want: "-"},
	}
	return nil
}
```
