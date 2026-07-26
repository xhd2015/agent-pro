# Agent Config Import/Export Tests

These doc-style tests verify the `pkgs/agentconfig` package, which provides
`Export` and `Import` functions for opencode, pi, and crush agent config files.
The tests use a synthetic home directory and never touch real user config.

## Decision Tree

```
operation?
├── export
│   ├── opencode-all/         all 4 source types present → zip has correct entries
│   ├── pi-all/               all 3 config files present → zip has correct entries
│   ├── crush-all/            both config files present → zip has correct entries
│   ├── opencode-missing/     no source directories → empty zip, no error
│   ├── pi-missing/           no pi directory → empty zip, no error
│   ├── crush-missing/        no crush directories → empty zip, no error
│   └── zip-create-error/     read-only zip directory → error
│
└── import
    ├── opencode-restore/     valid opencode zip → files restored
    ├── pi-restore/           valid pi zip → files restored
    ├── crush-restore/        valid crush zip → files restored
    ├── sensitive-0600/       auth/secret files → 0600 mode
    ├── non-sensitive-0644/   regular files → 0644 mode
    ├── path-traversal/       zip entry has "../" → rejected
    ├── unknown-prefix/       zip entry is "cursor/" → skipped
    ├── not-found/            zip file missing → error
    └── not-a-zip/            file is not a zip → error
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `export/opencode-all` | All 4 opencode sources (data json, opencode.jsonc, plugins, skills) → zip has correct entries |
| 2 | `export/pi-all` | All 3 pi files (auth.json, settings.json, models.json) → zip has correct entries |
| 3 | `export/crush-all` | Both crush files (config, data) → zip has correct entries |
| 4 | `export/opencode-missing` | No opencode directories exist → empty zip, no error |
| 5 | `export/pi-missing` | No pi directory exists → empty zip, no error |
| 6 | `export/crush-missing` | No crush directories exist → empty zip, no error |
| 7 | `export/zip-create-error` | Zip path in unwritable directory → error |
| 8 | `import/opencode-restore` | Valid zip with opencode entries → files restored to correct locations |
| 9 | `import/pi-restore` | Valid zip with pi entries → files restored to correct locations |
| 10 | `import/crush-restore` | Valid zip with crush entries → files restored to correct locations |
| 11 | `import/sensitive-0600` | Sensitive files (auth.json, opencode.jsonc, crush config) → 0600 mode |
| 12 | `import/non-sensitive-0644` | Non-sensitive files (settings.json, models.json, data) → 0644 mode |
| 13 | `import/path-traversal` | Zip entry with `../` in path → rejected, no escape |
| 14 | `import/unknown-prefix` | Zip entry with `cursor/` prefix → silently skipped |
| 15 | `import/not-found` | Zip file does not exist → error |
| 16 | `import/not-a-zip` | File exists but is not a valid zip → error |

## How to Run

```sh
# Vet the tree structure (no compilation):
doctest vet ./cmd/agent-pro/tests/config-import-export

# Build and run (will fail with RED until pkgs/agentconfig is implemented):
doctest test -v ./cmd/agent-pro/tests/config-import-export
```

```go
import (

	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	agentconfig "github.com/xhd2015/agent-pro/pkgs/agentconfig"
	"github.com/xhd2015/doctest/session"
)


type Request struct {
	Agent     string // "opencode", "pi", "crush"
	Operation string // "export", "import"
	ZipPath   string
	HomeDir   string
}

type Response struct {
	Err error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	var err error
	switch req.Operation {
	case "export":
		err = agentconfig.Export(req.Agent, req.HomeDir, req.ZipPath)
	case "import":
		err = agentconfig.Import(req.HomeDir, req.ZipPath)
	default:
		return nil, fmt.Errorf("unknown operation: %s", req.Operation)
	}
	return &Response{Err: err}, nil
}
```
