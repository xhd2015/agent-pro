# sessions.ListSessionFiles Tests

Doc-style tests for full session directory file listing in
`github.com/xhd2015/agent-pro/agent/grok/sessions`. Covers
`ListSessionFiles` and format helpers for CLI
`agent-pro grok session files <id> [--json]`. **Classic TDD** — RED until
implementer lands.

# DSN (Domain Specific Notion)

List every regular file in a Grok session directory (readdir of the session
folder under `sessions/<encoded-cwd>/<id>/`).

**Participants**

- **Caller** — CLI `session files` or in-process client needing session
  artifact inventory (summary, updates, signals, …).
- **`ListSessionFiles`** —
  `ListSessionFiles(grokHome, sessionID string) (dir string, files []SessionFile, err error)`.
  Locates session via same Find rules as Info/Status. Unknown session → error
  containing `grok session not found` and the id. On success, `dir` is the
  absolute session directory path; `files` is a readdir of that directory
  (regular files; skip subdirs if any).
- **`SessionFile`** — `Name` (basename), `Size` (bytes), `Mtime` (`time.Time`),
  `Path` (absolute path to the file).
- **Formatters**:
  - `FormatSessionFilesTable(files []SessionFile) string` — human table with
    columns **NAME**, **SIZE**, **MTIME** (stable header).
  - `FormatSessionFilesJSON(files []SessionFile) (string, error)` — JSON array
    of objects with at least `name`, `size`, `mtime` (RFC3339 or epoch
    acceptable if documented; tests accept RFC3339), optional `path`; no ANSI.

**Behaviors**

```
# list pipeline
grokHome + sessionID
  -> Find session dir
  -> missing -> error "grok session not found: <id>"
  -> readdir session dir -> []SessionFile (Name, Size, Mtime, Path)
  -> return (dir, files, nil)

# format
files -> FormatSessionFilesTable | FormatSessionFilesJSON
```

**Locked types**

```text
SessionFile
  Name  string
  Size  int64
  Mtime time.Time
  Path  string

ListSessionFiles(grokHome, sessionID string) (dir string, files []SessionFile, err error)
FormatSessionFilesTable(files []SessionFile) string
FormatSessionFilesJSON(files []SessionFile) (string, error)
```

## Version

0.0.2

## Decision Tree

```
agent/grok/sessions/tests/files/
├── DOCTEST.md
├── SETUP.md
├── multi-file/           # several files → all names listed with size/mtime/path
├── unknown-session/      # missing id → error
├── format-table/         # NAME SIZE MTIME header + rows
└── format-json/          # JSON array of file objects; no ANSI
```

Parameter ranking:

1. **Outcome** — success listing vs unknown session error
2. **Format** — structured list vs table vs json

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `multi-file` | Fixture with `summary.json`, `updates.jsonl`, `signals.json` → all three names present; dir path is session dir; sizes ≥ 1 |
| 2 | `unknown-session` | No session for id → error `grok session not found` + id |
| 3 | `format-table` | Multi-file fixture → table header includes NAME, SIZE, MTIME; each basename appears |
| 4 | `format-json` | Multi-file fixture → JSON array; each object has name; no ANSI |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/files
doctest test ./agent/grok/sessions/tests/files
doctest test -v ./agent/grok/sessions/tests/files/multi-file
```

Classic TDD: leaves RED until `ListSessionFiles` and format helpers exist.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	GrokHome  string
	TempDir   string
	SessionID string
	// Format: "" structured only; "table" | "json"
	Format string
}

type Response struct {
	Dir    string
	Files  []sessions.SessionFile
	Output string
	Err    error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	dir, files, err := sessions.ListSessionFiles(req.GrokHome, req.SessionID)
	resp := &Response{Dir: dir, Files: files, Err: err}
	if err != nil {
		return resp, nil
	}
	switch req.Format {
	case "table":
		resp.Output = sessions.FormatSessionFilesTable(files)
	case "json":
		out, jerr := sessions.FormatSessionFilesJSON(files)
		if jerr != nil {
			resp.Err = jerr
			return resp, nil
		}
		resp.Output = out
	}
	return resp, nil
}

var _ = time.Time{}
```
