package agentruncli

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"time"

	frontend "github.com/xhd2015/agent-pro/frontend-agent-run"
	"github.com/xhd2015/agent-pro/pkgs/assets"
)

// FrontendSource names where the SPA tree came from.
type FrontendSource string

const (
	FrontendSourceEmbed    FrontendSource = "embed"
	FrontendSourceCache    FrontendSource = "cache"
	FrontendSourceDownload FrontendSource = "download"
	FrontendSourceA1       FrontendSource = "a1"
)

// ResolvedFrontend is the result of resolveAgentRunFrontend.
// When Source is FrontendSourceA1, FS is nil and callers should serve the A1 shell.
type ResolvedFrontend struct {
	FS       fs.FS
	Source   FrontendSource
	CacheDir string
	// EnsureErr is set when download was attempted and failed (A1 fallback).
	EnsureErr error
}

// resolveAgentRunFrontend picks a complete SPA tree for agent-run:
//
//  1. complete embedded dist (fat build) — offline, no network
//  2. complete on-disk asset cache
//  3. EnsureAsset download into cache (uses EnsureConfig / AGENT_PRO_ASSET_BASE_URL)
//  4. A1 incomplete shell (FS nil)
func resolveAgentRunFrontend(ctx context.Context, cfg assets.EnsureConfig) ResolvedFrontend {
	return resolveProductFrontend(ctx, assets.ProductAgentRun, frontend.DistComplete, frontend.DistFS, cfg)
}

func resolveProductFrontend(
	ctx context.Context,
	product string,
	distComplete func() bool,
	distFS fs.FS,
	cfg assets.EnsureConfig,
) ResolvedFrontend {
	version := assets.ClientVersion()

	if distComplete() {
		sub, err := fs.Sub(distFS, "dist")
		if err == nil {
			return ResolvedFrontend{FS: sub, Source: FrontendSourceEmbed}
		}
	}

	if assets.CacheComplete(product, version, assets.KindFrontend) {
		dir, err := assets.AssetCacheDir(product, version, assets.KindFrontend)
		if err == nil {
			return ResolvedFrontend{FS: os.DirFS(dir), Source: FrontendSourceCache, CacheDir: dir}
		}
	}

	if ctx == nil {
		ctx = context.Background()
	}
	dir, err := assets.EnsureAsset(ctx, product, version, assets.KindFrontend, cfg)
	if err == nil {
		return ResolvedFrontend{FS: os.DirFS(dir), Source: FrontendSourceDownload, CacheDir: dir}
	}

	return ResolvedFrontend{Source: FrontendSourceA1, EnsureErr: err}
}

// a1IncompleteHTML is a minimal operator-facing page when no complete SPA is available.
func a1IncompleteHTML(product string) string {
	version := assets.ClientVersion()
	env := assets.EnvAssetBaseURL
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>%s — frontend assets incomplete</title>
  <style>
    body { font-family: system-ui, sans-serif; max-width: 42rem; margin: 2rem auto; padding: 0 1rem; line-height: 1.5; color: #1a1a1a; }
    code { background: #f4f4f4; padding: 0.1em 0.35em; border-radius: 4px; }
    pre { background: #f4f4f4; padding: 0.75rem 1rem; border-radius: 6px; overflow-x: auto; }
    h1 { font-size: 1.25rem; }
  </style>
</head>
<body>
  <h1>Frontend assets incomplete</h1>
  <p>
    This <code>%s</code> binary does not include a complete embedded UI
    (version <code>%s</code>), and no complete asset cache was found.
  </p>
  <p>To fix:</p>
  <ol>
    <li>Build a fat binary with a full frontend dist, or</li>
    <li>
      Set <code>%s</code> to your release asset host and run:
      <pre>agent-run assets ensure</pre>
      then restart the server.
    </li>
  </ol>
  <p>
    Cache layout: <code>$XDG_CACHE_HOME/agent-pro/asset-cache/%s/%s/frontend</code>
    (or <code>~/.cache/...</code>).
  </p>
</body>
</html>
`, product, product, version, env, product, version)
}

// defaultEnsureContext is a short timeout for non-blocking startup ensure attempts.
func defaultEnsureContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
