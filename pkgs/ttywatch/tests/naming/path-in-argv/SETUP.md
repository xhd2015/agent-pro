# Scenario

**Feature**: filesystem paths in argv are sanitized to underscore tokens

```
argv [/usr/local/bin/codex, C:\Program Files\Grok\grok.exe]
  -> SlugifyCommandLine -> no raw / or \; collapsed underscores
  -> ServeSubcommand -> __serve_{slug}__
```

## Steps

1. Set `req.Argv` with Unix absolute path and Windows-style path segments.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Argv = []string{
		"/usr/local/bin/codex",
		`C:\Program Files\Grok\grok.exe`,
	}
	return nil
}
```