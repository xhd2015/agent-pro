# Timeout Duration Parsing Tests

Verify the exported function `subagent.ParseTimeoutDuration(s string) (time.Duration, error)` — the
parser for the `--timeout` flag used by `doctest agent design` and `doctest agent implement`.

## Behavior Summary

| Input | Expected Behavior |
|---|---|
| `""` (empty) | Return `1h`, nil error, no stderr |
| Bare number (no letters) | Treated as seconds, then parsed with `time.ParseDuration` |
| Valid Go duration string | Parsed with `time.ParseDuration` |
| Whitespace | Trimmed before parsing |
| Duration < 1m | Error (must be at least 1 minute) |
| 1m ≤ duration < 10m | Warning to stderr suggesting longer timeout |
| Duration ≥ 10m | Clean parse, no warning |
| Invalid input | Parse error |

## Decision Tree

```
agent/subagent/tests/timeout/
├── DOCTEST.md                          # This file
├── SETUP.md                            # Root: Request{Input}, Response{Duration,Stderr,Err}
│
├── valid/                              # === Valid inputs (no error, no warning) ===
│   ├── SETUP.md
│   ├── default-empty/                  # Input: ""  → Duration=1h
│   ├── suffix-hours/                   # Input: "1h" → Duration=1h
│   ├── combined/                       # Input: "1h30m" → Duration=1h30m
│   └── valid-10m-boundary/             # Input: "10m" → Duration=10m (warning threshold)
│
├── bare-number/                        # === Bare number (no suffix) ===
│   ├── SETUP.md
│   └── bare-seconds/                   # Input: "30" → interpreted as 30s → error (< 1m)
│
├── warning/                            # === Warning range: 1m ≤ d < 10m ===
│   ├── SETUP.md
│   ├── suffix-minutes/                 # Input: "5m" → Duration=5m, warning on stderr
│   ├── warning-less-than-10m/          # Input: "3m" → Duration=3m, warning on stderr
│   └── with-whitespace/                # Input: "  5m  " → Duration=5m (whitespace trimmed)
│
└── error/                              # === Error cases ===
    ├── SETUP.md
    ├── suffix-seconds/                 # Input: "30s" → error (< 1m)
    ├── error-less-than-1m/            # Input: "30s" → error, message mentions "at least 1m"
    └── error-invalid/                  # Input: "abc" → error (parse failure)
```

The primary split is by **outcome category** (valid, bare-number, warning, error), since the single
input string directly determines the outcome. Within each category, individual leaves test specific
values at boundaries and representative points.

## Test Index

### valid — 4 leaves
| Leaf | Input | Duration |
|------|-------|----------|
| `default-empty` | `""` | `1h` |
| `suffix-hours` | `"1h"` | `1h` |
| `combined` | `"1h30m"` | `1h30m` |
| `valid-10m-boundary` | `"10m"` | `10m` |

### bare-number — 1 leaf
| Leaf | Input | Result |
|------|-------|--------|
| `bare-seconds` | `"30"` | error (< 1m) |

### warning — 3 leaves
| Leaf | Input | Duration | Stderr |
|------|-------|----------|--------|
| `suffix-minutes` | `"5m"` | `5m` | warning |
| `warning-less-than-10m` | `"3m"` | `3m` | warning |
| `with-whitespace` | `"  5m  "` | `5m` | warning |

### error — 3 leaves
| Leaf | Input | Result |
|------|-------|--------|
| `suffix-seconds` | `"30s"` | error |
| `error-less-than-1m` | `"30s"` | error, mentions minimum |
| `error-invalid` | `"abc"` | error (parse failure) |

Total: **11 leaves** across **4 grouping nodes**.

## Coverage Checklist

- [x] Default/empty input
- [x] Bare number (no suffix → seconds)
- [x] All standard Go duration suffixes (s, m, h)
- [x] Combined duration (`1h30m`)
- [x] Whitespace trimming
- [x] Error: duration < 1m (bare and suffix)
- [x] Warning: 1m ≤ duration < 10m (3m, 5m)
- [x] Error: invalid input
- [x] Boundary: exactly 10m (warning threshold)

## How to Run

```sh
doctest test -v ./external/agent-pro/agent/subagent/tests/timeout
```

Or from the repo root:

```sh
doctest test ./agent/subagent/tests/timeout
```
