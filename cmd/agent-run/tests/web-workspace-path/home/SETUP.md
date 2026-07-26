# Scenario

**Feature**: home page (`/`) WorkspacePath — status.workspace + runner chrome

```
# open API home surface
agent-run web (no --token) -> GET / -> status.workspace drives WorkspacePath
  -> runner-picker must stay in 390px viewport when collapsed
```

## Preconditions

- Open API mode (`WebTokenMode=omit`) so auth page is skipped without localStorage token.
- Home leaves open `/` and wait for `[data-testid="workspace"]`.

## Steps

1. Set open API token mode and clear any default explicit token expectation.
2. Descendant leaves set `WebWorkingDir` / script.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WebTokenMode = "omit"
	req.Token = ""
	return nil
}
```
