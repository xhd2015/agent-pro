# Scenario

**Feature**: ttywatch derives stable `__serve_{slug}__` reexec tokens from run argv

```
# pure functions: argv -> SlugifyCommandLine -> ServeSubcommand
leaf Setup sets req.Argv -> Run calls pkgs/ttywatch naming helpers
doctest <- Response.Slug and Response.ServeSubcommand
```

## Preconditions

- `pkgs/ttywatch/naming.go` implements `SlugifyCommandLine` and `ServeSubcommand` (added by implementer).

## Steps

1. Leaf `Setup` sets `req.Argv` for the scenario variant.
2. `Run` calls both naming helpers and returns slug + wrapped token.
3. Leaf `Assert` checks sanitization, wrapping, or truncation rules.

## Context

- These are unit-style doctests: no PTY, registry, or subprocess.
- `ServeSubcommand(argv)` must equal `__serve_` + `SlugifyCommandLine(argv)` + `__`.
- Truncated slugs must remain deterministic for a fixed argv fixture.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.Argv != nil {
		for i, arg := range req.Argv {
			if arg == "" {
				t.Fatalf("req.Argv[%d] must not be empty string", i)
			}
		}
	}
	return nil
}

// assertServeWrap checks ServeSubcommand wraps Slugify output with __serve_ / __.
func assertServeWrap(t *testing.T, slug, serve string) {
	t.Helper()
	want := "__serve_" + slug + "__"
	if serve != want {
		t.Fatalf("ServeSubcommand = %q, want %q (slug %q)", serve, want, slug)
	}
	if !strings.HasPrefix(serve, "__serve_") || !strings.HasSuffix(serve, "__") {
		t.Fatalf("serve token must match __serve_{slug}__ pattern, got %q", serve)
	}
	if serve == "__serve__" {
		t.Fatalf("bare __serve__ is no longer valid; got %q", serve)
	}
}

// assertNoBareServe rejects the legacy constant token.
func assertNoBareServe(t *testing.T, serve string) {
	t.Helper()
	if serve == "__serve__" {
		t.Fatal("legacy __serve__ token must not be produced")
	}
}
```