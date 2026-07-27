# Scenario

**Feature**: residual update banner alone must not block idle when model is not loading

```
04-idle-prompt.snapshot.txt with model:loading stripped -> IsBlocking=false, writable idle
```

## Preconditions

Base fixture `04-idle-prompt.snapshot.txt`. Harness sets `StripModelLoading=true`.

## Steps

1. `FixtureFile=04-idle-prompt.snapshot.txt`.
2. `StripModelLoading=true`.

## Context

Banner is non-blocking; still honor real `model:loading` (menu-dismissed leaf).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.FixtureFile = "04-idle-prompt.snapshot.txt"
	req.StripModelLoading = true
	return nil
}
```
