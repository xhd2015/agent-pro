# Codex Models Tests

Doc-style tests for `agent/codex/models`, which lists Codex CLI models from a
synthetic home (`config.toml` + `models_cache.json`) and formats them for
`agent-pro codex models` (human text and `--json`).

# DSN (Domain Specific Notion)

Codex caches a model catalog in `~/.codex/models_cache.json` with per-model
`slug`, `display_name`, `visibility`, `default_reasoning_level`, and
`supported_reasoning_levels`. Only `visibility == "list"` entries are shown.
`config.toml`'s top-level `model` becomes `Catalog.Default` and is unioned into
the list when the cache omits it.

CLI contract:
- human: `Home` / `Default` header, `* ` marks default, includes display name
  and reasoning; empty → `(no models)`
- `--json`: indented `Catalog` with unified model objects
  (`id`, `source`, `display_name?`, `default_reasoning?`, `reasoning?`)

## Version

0.0.1

## Decision Tree

```
format?
├── human/
│   ├── with-models/   list-visible only; default marked; reasoning shown
│   └── empty/         missing home files → "(no models)"
└── json/
    ├── with-models/   Catalog JSON with per-model reasoning
    └── empty/         models=[] and from_* false
```

Intentional exclusions: invalid-file errors and CODEX_HOME override (package
L1 unit tests); CLI flag parsing (thin wrapper).

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `human/with-models` | Human text marks default and shows reasoning; hides visibility≠list |
| 2 | `human/empty` | Empty home → `(no models)` |
| 3 | `json/with-models` | JSON Catalog includes reasoning; excludes hidden slugs |
| 4 | `json/empty` | JSON empty catalog |

## How to Run

```sh
cd external/agent-pro
doctest vet ./cmd/agent-pro/tests/codex-models
doctest test -v ./cmd/agent-pro/tests/codex-models
```

```go
import (
	"fmt"
	"testing"

	codexmodels "github.com/xhd2015/agent-pro/agent/codex/models"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	CodexHome string
	Format    string // "human" or "json"
}

type Response struct {
	Catalog codexmodels.Catalog
	Output  string
	JSON    []byte
	Err     error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	resp := &Response{}
	cat, err := codexmodels.List(req.CodexHome)
	if err != nil {
		resp.Err = err
		return resp, nil
	}
	resp.Catalog = cat
	switch req.Format {
	case "json":
		data, err := codexmodels.FormatJSON(cat)
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		resp.JSON = data
	case "human", "":
		resp.Output = codexmodels.FormatText(cat)
	default:
		resp.Err = fmt.Errorf("unknown format: %s", req.Format)
	}
	return resp, nil
}
```
