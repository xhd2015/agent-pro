# Scenario

**Feature**: single --grep keeps only matching sessions; title uses match totals

```
# 3 sessions (k8s, docker, redis); explain list --grep kubernetes
-> only k8s card; title 1 shown of 1, limit 10; docker/redis absent
```

## Preconditions

- Three distinctive sessions from `grepTripleSessions`.

## Steps

1. Seed triple fixture.
2. Args: `list --grep kubernetes`.
3. Assert only k8s markers; title match totals; non-matches absent.

## Context

- Core single-pattern filter path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "kubernetes"}
	req.Sessions = grepTripleSessions()
	return nil
}
```
