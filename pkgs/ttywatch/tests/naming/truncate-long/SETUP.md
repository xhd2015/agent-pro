# Scenario

**Feature**: oversized argv slugs truncate to ~120 chars with deterministic hash suffix

```
argv [codex + 40 flag pairs]
  -> SlugifyCommandLine -> len <= ~128 with _<hash> suffix
  -> ServeSubcommand -> __serve_{slug}__ (not bare __serve__)
```

## Steps

1. Build a deliberately long argv with many `--flag-N value-N` pairs.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	argv := []string{"codex"}
	for i := 0; i < 40; i++ {
		argv = append(argv,
			fmt.Sprintf("--flag-%02d", i),
			fmt.Sprintf("value-with-many-chars-number-%02d", i),
		)
	}
	req.Argv = argv
	return nil
}
```