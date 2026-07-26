# Scenario

**Feature**: exact production symptom is classified as transient

```
IsTransientIndexError("fatal: unable to write new index file", nil) -> true
```

## Preconditions

- Input is the exact production git fatal line (case as emitted by git).

## Steps

1. Set `ClassifyOutput` to `fatal: unable to write new index file`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClassifyOutput = "fatal: unable to write new index file"
	req.WantTransient = true
	return nil
}
```
