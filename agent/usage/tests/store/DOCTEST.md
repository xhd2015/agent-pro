# usage snapshot store tests

Doc-style tests for the on-disk snapshot store in
`github.com/xhd2015/agent-pro/agent/usage`: `Store.Append` writes one JSON line
per collect cycle and `Store.List` reads them back newest first.

# DSN (Domain Specific Notion)

`agent-pro usage collect` appends one snapshot per provider under

```
$AGENT_PRO_HOME/usages/<provider>/<YYYY-MM-DD>/<HH-MM-SS>-snapshot.jsonl
```

**Participants**

- **Store** — `usage.NewStore(root)`; `Append(record)` returns the file it wrote,
  `List(last)` returns up to `last` stored records (0 = all) newest first;
  reading stops on an instant boundary, so truncation can never return an older
  snapshot in place of a newer one from another provider.
- **Snapshot file** — JSON Lines, one record per collect cycle for one provider.
  Directory names are local time so the tree is browsable with `ls`; the `ts`
  inside each record is UTC, so records compare across machines.
- **Same-second runs** — a second cycle inside the same second appends another
  line to the same file; a run is never overwritten, and the file is the unit a
  reader pages through.

```
Append(record) -> usages/<provider>/<local date>/<local time>-snapshot.jsonl
                  + "{\"schema\":...,\"ts\":\"<UTC>\",...}\n"   (append, never truncate)
List(last)     -> newest-first records across every provider and day
doctest <- written paths, raw file lines, records read back
```

## Version

0.0.1

## Decision Tree

```
agent/usage/tests/store/
├── DOCTEST.md
├── SETUP.md
├── append/
│   ├── writes-layout/        # path layout, one UTC record per line, round trip
│   └── same-second-appends/  # two cycles in one second -> one file, two lines
└── list/
    ├── newest-first/         # ordering across providers and days, --last truncation
    ├── no-snapshots/         # a root with unrelated files holds no snapshots
    └── missing-root/         # a root that was never created is not an error
```

Parameter ranking (most → least significant):

1. **Write semantics** — append vs overwrite when a file already exists
2. **Layout** — provider directory, local date and time in the path, UTC `ts` inside
3. **Read order** — newest first, ties by provider
4. **Truncation** — `last` limits how many records come back, newest instants first
5. **Absent or unrelated roots** — empty result, no error

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `append/writes-layout` | Two providers land in `<provider>/<date>/<time>-snapshot.jsonl`, one UTC line each |
| 2 | `append/same-second-appends` | Three cycles in the same second share one file, in order |
| 3 | `list/newest-first` | Cross-provider ordering plus `last` truncation |
| 4 | `list/no-snapshots` | A root holding unrelated files yields no records |
| 5 | `list/missing-root` | A root that does not exist yields no records and no error |

## How to Run

```sh
doctest vet ./agent/usage/tests/store
doctest test -v ./agent/usage/tests/store
doctest test -v ./agent/usage/tests/store/append/same-second-appends
```

```go
import (
"bufio"
"encoding/json"
"fmt"
"os"
"path/filepath"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

type Request struct {
// Root is the usages root; it may point at a path that does not exist.
Root string
// Appends are the records written by Run, in order.
Appends []usage.Record
// ListLasts are the Store.List arguments Run uses after appending; an empty
// slice skips listing.
ListLasts []int
}

type Response struct {
// Paths are the files the appends landed in, in append order.
Paths []string
// Lines holds the raw lines of every file written.
Lines map[string][]string
// Stored holds one Store.List result per entry in ListLasts.
Stored [][]usage.StoredRecord
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
t.Helper()
store := usage.NewStore(req.Root)
resp := &Response{Lines: map[string][]string{}}

for _, rec := range req.Appends {
path, err := store.Append(rec)
if err != nil {
return resp, err
}
resp.Paths = append(resp.Paths, path)
}

seen := map[string]bool{}
for _, path := range resp.Paths {
if seen[path] {
continue
}
seen[path] = true
lines, err := ReadLines(path)
if err != nil {
return resp, err
}
resp.Lines[path] = lines
}

for _, last := range req.ListLasts {
stored, err := store.List(last)
if err != nil {
return resp, err
}
resp.Stored = append(resp.Stored, stored)
}
return resp, nil
}

// Snapshot builds one stored record for a provider at a fixed instant.
func Snapshot(provider usage.ProviderID, ts time.Time, usedPercent int) usage.Record {
return usage.Record{
Schema:   usage.SchemaID,
TS:       ts.UTC().Truncate(time.Second),
Provider: provider,
Usage: usage.UsageBlock{
OK:         true,
Endpoint:   "billing",
DurationMS: 7,
Values:     map[string]float64{"used_percent": float64(usedPercent)},
Display:    map[string]string{"usage": fmt.Sprintf("%d%% used", usedPercent)},
},
Sessions: usage.SessionsBlock{Total: usedPercent},
}
}

// ReadLines returns the lines of one snapshot file, without the newline.
func ReadLines(path string) ([]string, error) {
file, err := os.Open(path)
if err != nil {
return nil, err
}
defer file.Close()

var lines []string
scanner := bufio.NewScanner(file)
for scanner.Scan() {
if line := scanner.Text(); line != "" {
lines = append(lines, line)
}
}
return lines, scanner.Err()
}

// OnDisk is the record one stored file line holds, decoded independently of the
// usage.Record type so a leaf can assert the wire shape itself.
type OnDisk struct {
Schema   string `json:"schema"`
TS       string `json:"ts"`
Provider string `json:"provider"`
Usage    struct {
OK       bool               `json:"ok"`
Endpoint string             `json:"endpoint"`
Values   map[string]float64 `json:"values"`
Display  map[string]string  `json:"display"`
} `json:"usage"`
Sessions struct {
Total  int     `json:"total"`
Oldest *string `json:"oldest"`
Newest *string `json:"newest"`
} `json:"sessions"`
}

// DecodeLine decodes one snapshot file line.
func DecodeLine(t *testing.T, line string) OnDisk {
t.Helper()
var record OnDisk
if err := json.Unmarshal([]byte(line), &record); err != nil {
t.Fatalf("decode snapshot line: %v\n%s", err, line)
}
return record
}

// LayoutPath is the path a record is expected to be written to: the provider
// directory, the local date, and the local time with the snapshot suffix.
func LayoutPath(root string, rec usage.Record) string {
local := rec.TS.Local()
return filepath.Join(root, string(rec.Provider),
local.Format("2006-01-02"), local.Format("15-04-05")+usage.SnapshotSuffix)
}
```
