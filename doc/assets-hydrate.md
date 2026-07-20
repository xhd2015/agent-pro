# Asset hydrate

How **agent-pro** and **agent-run** obtain SPA frontends when the binary was
built without a full fat embed.

## Why placeholders exist

Git tracks thin placeholder `dist/` trees (or empty embeds) so the module stays
small and CI can build without bundling JS. Fat Vite output under:

```text
frontend/dist/                 # agent-pro SPA
frontend-agent-run/dist/       # agent-run SPA
```

is produced by local/install builds and is not the source of truth for release
downloads. Operators who `go install` or build with incomplete embeds get a
**thin** binary. Hydrate fills the gap from versioned release archives without
committing built assets.

Completeness = non-empty `index.html` at dist root **and** at least one
non-empty file under `assets/` (`frontend.DistComplete` /
`frontend-agent-run.DistComplete`, same rules as the on-disk cache).

## Fat install / build scripts

Stage fat dist, then install:

```bash
# agent-pro (recommended): both SPAs + binary
#   frontend/              traces viewer
#   frontend-agent-run/    grok session view --web (message cards)
go run ./script/agent-pro/bundle    # fat both dist trees only
go run ./script/agent-pro/install   # bundle + go install ./cmd/agent-pro

# legacy alias (delegates to script/agent-pro/install)
go run ./script/install

# agent-run only
go run ./script/agent-run/build-frontend
go run ./script/agent-run/install

# low-level SPA builders (used by the scripts above)
go run ./script/build-frontend              # frontend/
go run ./script/agent-run/build-frontend    # frontend-agent-run/
```

After a fat dist is present, `//go:embed` ships a complete SPA and runtime
hydrate is unnecessary for that product.

Vite empties `dist/` during build. The frontend build scripts rewrite
`dist/placeholder.txt` afterward so the tracked thin-embed marker stays on disk
for git / clean checkouts.

## Cache layout + base URL

Downloaded archives unpack under:

```text
# Prefer XDG when set:
$XDG_CACHE_HOME/agent-pro/asset-cache/{product}/{version}/{kind}/

# Default when XDG_CACHE_HOME is unset:
~/.cache/agent-pro/asset-cache/{product}/{version}/{kind}/
```

Examples (version normalized with leading `v`):

- `~/.cache/agent-pro/asset-cache/agent-run/v0.0.70/frontend/`
- `~/.cache/agent-pro/asset-cache/agent-pro/v0.0.70/frontend/`

Environment:

- `AGENT_PRO_ASSET_BASE_URL` — download base (no trailing slash required).  
  Archive URL:
  `{BaseURL}/v{version}/{product}_v{version}_frontend.tar.gz`  
  e.g. `…/v0.0.70/agent-run_v0.0.70_frontend.tar.gz`  
  (never the `latest` tag).

## Runtime resolution

1. If the live embed is **complete** → serve embed (no download).
2. Else if the local **asset-cache** is complete → serve cache.
3. Else **download** via `assets.EnsureAsset` when BaseURL is set; otherwise
   surface a clear incomplete-embed error.

## Operator CLI (`assets status` / `assets ensure`)

**agent-run**:

```bash
agent-run assets status   # embed + cache completeness (no network)
agent-run assets ensure   # download into cache if needed
agent-run assets --help
```

`assets ensure` uses `AGENT_PRO_ASSET_BASE_URL`, writes under the cache layout
above, and skips the network when embed or cache is already complete.

agent-pro web serving uses the same `pkgs/assets` ensure path for thin embeds
(see server frontend resolve).

## Release archive names

Basenames must match `assets.AssetReleaseNames(version)` / `AssetArchiveName`:

- `agent-run_v0.0.70_frontend.tar.gz`
- `agent-pro_v0.0.70_frontend.tar.gz`

Archive **root** = dist contents (`index.html` at tar root, then `assets/…`).

## Publishing release assets

From the agent-pro module root, pack hydrate archives with
**`script/github/release-assets`**:

```bash
# Pack only (no network / no gh)
go run ./script/github/release-assets --out ./dist --version v0.0.70

# Default --out is a temp dir; prints out: <abs-path>
go run ./script/github/release-assets --version v0.0.70
```

Sources:

- `frontend-agent-run/dist` → `agent-run_…_frontend.tar.gz`
- `frontend/dist` → `agent-pro_…_frontend.tar.gz`

Incomplete (thin) dist fails with a clear error. Optional **`--upload`** wraps
**`gh`**: create the release tag if missing, then `gh release upload --clobber`:

```bash
go run ./script/github/release-assets --out ./dist --version v0.0.70 --upload
```

Default is pack-only; use `--upload` only when publishing (requires `gh` on
`PATH` and an authenticated repo). macOS: `COPYFILE_DISABLE=1` and `._*` /
`.DS_Store` are excluded from archives.

## Summary

| Situation | Behavior |
|-----------|----------|
| Fat install/embed | Offline; serve from embed |
| Thin embed + complete cache | Serve cache |
| Thin embed + incomplete cache | Download via `AGENT_PRO_ASSET_BASE_URL` |
| Operator control | `agent-run assets status` / `assets ensure` |
| Publish hydrate archives | `go run ./script/github/release-assets` (+ `--upload` for gh) |
