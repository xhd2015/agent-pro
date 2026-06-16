# Config Import/Export Integration Tests (Podman)

These integration tests verify the config export/import cycle end-to-end using
a podman Linux container. Real configs are exported from the host, copied into
a Debian container, extracted to correct agent paths, and then the agent binary
is queried to confirm the configs are available.

Each leaf covers one agent (opencode, pi, or crush) with its specific config
paths, binary installation method, and query format. Skip conditions are
checked before any container is created (podman available, configs exist,
binary obtainable).

## Decision Tree

```
agent type?
├── opencode
│   ├── podman available? ── no ── SKIP
│   ├── configs exist? ── no ── SKIP
│   ├── binary obtainable? ── no ── SKIP
│   └── query returns "paris" ── opencode-podman leaf
├── pi
│   ├── podman available? ── no ── SKIP
│   ├── configs exist? ── no ── SKIP
│   ├── binary obtainable? ── no ── SKIP
│   └── query returns "paris" ── pi-podman leaf
└── crush
    ├── podman available? ── no ── SKIP
    ├── configs exist? ── no ── SKIP
    ├── binary obtainable? ── no ── SKIP
    └── query returns "paris" ── crush-podman leaf
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `opencode-podman` | Export opencode configs, run in container with `opencode run --format json "one word of French capital"` → output contains "paris" |
| 2 | `pi-podman` | Export pi configs, run in container with `pi -p "one word of French capital" --mode json --approve` → output contains "paris" |
| 3 | `crush-podman` | Export crush configs, run in container with `crush run --verbose "one word of French capital"` → output contains "paris" |

## How to Run

```sh
# Vet the tree structure:
doctest vet ./cmd/agent-pro/tests/config-import-export/integration

# Build and run (may skip or fail depending on host environment):
doctest test -v ./cmd/agent-pro/tests/config-import-export/integration
```
