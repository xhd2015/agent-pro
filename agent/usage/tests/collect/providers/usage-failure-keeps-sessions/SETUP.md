# Scenario

**Feature**: a failed usage fetch still produces a snapshot that carries the session counts

```
fixtures removed + no Command Code credentials
Collect -> grok error, codex error, commandcode error
        -> usage{ok:false, error} + sessions{total, oldest, newest} for every provider
```

## Preconditions

- `Fixtures` is a non-nil empty map, so the fixture transport answers every grok
  and codex URL with a missing file instead of reaching the network.
- `CommandCodeCredentialed` is false, so the collect uses the credential-less home
  and fails locally.
- Each home still has a session, including the credential-less Command Code home:
  counting runs whether or not the fetch worked.

## Steps

1. Drop every fixture and the Command Code credentials.
2. Seed a grok session and a Command Code transcript that the fetch cannot affect.

```go
import (
"testing"
"time"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Fixtures = map[string]string{}
req.CommandCodeCredentialed = false

WriteGrokSession(t, req.GrokHome, "2026-09-11T09-00-00-main",
"2026-09-11T09:00:00Z", "2026-09-11T10:00:00Z")
WriteCommandCodeTranscript(t, req.BareCommandCodeHome, "-Users-tester-project",
"33333333-3333-4333-8333-333333333333.jsonl",
`{"type":"session","timestamp":"2026-09-11T09:00:00Z"}`,
req.Now.Add(-24*time.Hour))
return nil
}
```
