# Scenario

**Feature**: per-session `--head` / `--tail` slice with omission markers

```
# after text filters, chrono prompts
head N -> first N; OmittedAfter = total-N if total>N; marker AFTER lines
tail N -> last N; OmittedBefore = total-N if total>N; marker BEFORE lines
```

## Preconditions

- HeadSet XOR TailSet (errors covered under filter/errors).
- N >= 1 when set.
- Marker only when M > 0; not a UserPrompt.
- Unit is per-session (not global).

## Steps

1. Seed session(s) with known chrono prompt lists.
2. Set Head or Tail.
3. Assert structured Omitted* and/or format marker placement; footer counts.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	_ = req
	return nil
}
```
