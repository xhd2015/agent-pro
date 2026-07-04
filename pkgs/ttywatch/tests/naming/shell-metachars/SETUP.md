# Scenario

**Feature**: shell metacharacters in argv become underscore separators

```
argv [sh -c echo;whoami|cat&done]
  -> SlugifyCommandLine -> metachars ; | & $ sanitized
  -> ServeSubcommand -> __serve_{slug}__
```

## Steps

1. Set `req.Argv` with a `sh -c` script fragment containing common metacharacters.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Argv = []string{"sh", "-c", "echo;whoami|cat&done$HOME"}
	return nil
}
```