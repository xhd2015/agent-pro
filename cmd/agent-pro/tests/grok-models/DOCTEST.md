# Grok Models Tests

Doc-style tests for `agent/grok/models`, which lists Grok CLI models from a
synthetic home (`config.toml` + `models_cache.json`) and formats them for
`agent-pro grok models` (human text and `--json`).

# DSN (Domain Specific Notion)

Grok stores model ids in `~/.grok/config.toml` under `[model."…"]` plus an
optional `[models].default`, and may also cache ids in `models_cache.json`.
`List` unions both sources (sorted), soft-fails when files are missing, and
errors when a present file is invalid.

CLI contract:
- human: `Home` / `Default` header, `* ` marks the configured default, optional
  display name, empty → `(no models)`
- `--json`: indented `Catalog` with unified model objects
  (`id`, `source`, `display_name?`) plus `home` / `default` / `from_*`

## Version

0.0.1

## Decision Tree

```
format?
├── human/
│   ├── with-models/   config+cache union; default marked with *
│   └── empty/         missing home files → "(no models)"
└── json/
    ├── with-models/   Catalog JSON with sorted models + default
    └── empty/         models=[] and from_* false
```

Intentional exclusions: invalid-file errors (covered by package L1 unit tests);
CLI flag parsing (thin wrapper over List + Format*).

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `human/with-models` | Human text lists sorted ids and marks default with `*` |
| 2 | `human/empty` | Empty home → `(no models)` |
| 3 | `json/with-models` | JSON Catalog matches List fields |
| 4 | `json/empty` | JSON empty catalog |

## How to Run

```sh
cd external/agent-pro
doctest vet ./cmd/agent-pro/tests/grok-models
doctest test -v ./cmd/agent-pro/tests/grok-models
```

```go
import (
	"fmt"
	"testing"

	grokmodels "github.com/xhd2015/agent-pro/agent/grok/models"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	GrokHome string
	Format   string // "human" or "json"
}

type Response struct {
	Catalog grokmodels.Catalog
	Output  string
	JSON    []byte
	Err     error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	resp := &Response{}
	cat, err := grokmodels.List(req.GrokHome)
	if err != nil {
		resp.Err = err
		return resp, nil
	}
	resp.Catalog = cat
	switch req.Format {
	case "json":
		data, err := grokmodels.FormatJSON(cat)
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		resp.JSON = data
	case "human", "":
		resp.Output = grokmodels.FormatText(cat)
	default:
		resp.Err = fmt.Errorf("unknown format: %s", req.Format)
	}
	return resp, nil
}
```
