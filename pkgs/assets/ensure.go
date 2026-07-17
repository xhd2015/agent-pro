package assets

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// EnvAssetBaseURL is the environment variable for the default asset BaseURL.
const EnvAssetBaseURL = "AGENT_PRO_ASSET_BASE_URL"

// EnsureConfig configures EnsureAsset downloads.
type EnsureConfig struct {
	// BaseURL is the release asset host prefix, e.g. https://example.com/assets.
	// If empty, AGENT_PRO_ASSET_BASE_URL is used.
	BaseURL string
	// HTTPClient is optional; defaults to http.DefaultClient.
	HTTPClient *http.Client
}

// ResolveBaseURL returns cfg.BaseURL or the AGENT_PRO_ASSET_BASE_URL env value.
func ResolveBaseURL(cfg EnsureConfig) string {
	if s := strings.TrimSpace(cfg.BaseURL); s != "" {
		return strings.TrimRight(s, "/")
	}
	return strings.TrimRight(strings.TrimSpace(os.Getenv(EnvAssetBaseURL)), "/")
}

var (
	ensureMu    sync.Mutex
	ensureGroup = map[string]*ensureCall{}
)

type ensureCall struct {
	wg  sync.WaitGroup
	dir string
	err error
}

// EnsureAsset ensures the SPA for product/version/kind is present and complete
// under the asset cache. If the cache is already complete, no network is used.
// Concurrent calls for the same key single-flight.
//
// Archive URL:
//
//	{BaseURL}/v{version}/{product}_v{version}_{kind}.tar.gz
//
// The archive root is the dist tree (index.html at archive root).
// Returns the cache directory path on success.
func EnsureAsset(ctx context.Context, product, version, kind string, cfg EnsureConfig) (string, error) {
	v := NormalizeVersion(version)
	if product == "" || v == "" || kind == "" {
		return "", fmt.Errorf("ensure asset: product, version, and kind are required")
	}
	dir, err := AssetCacheDir(product, v, kind)
	if err != nil {
		return "", err
	}
	if spaDirComplete(dir) {
		return dir, nil
	}

	key := product + "\x00" + v + "\x00" + kind
	ensureMu.Lock()
	if c, ok := ensureGroup[key]; ok {
		ensureMu.Unlock()
		c.wg.Wait()
		return c.dir, c.err
	}
	c := &ensureCall{}
	c.wg.Add(1)
	ensureGroup[key] = c
	ensureMu.Unlock()

	c.dir, c.err = ensureAssetDownload(ctx, product, v, kind, dir, cfg)
	c.wg.Done()

	ensureMu.Lock()
	delete(ensureGroup, key)
	ensureMu.Unlock()

	return c.dir, c.err
}

func ensureAssetDownload(ctx context.Context, product, version, kind, destDir string, cfg EnsureConfig) (string, error) {
	// Re-check after winning the single-flight race.
	if spaDirComplete(destDir) {
		return destDir, nil
	}

	base := ResolveBaseURL(cfg)
	if base == "" {
		return "", fmt.Errorf("ensure asset: BaseURL is required (set EnsureConfig.BaseURL or %s)", EnvAssetBaseURL)
	}

	url := base + AssetReleaseURLPath(product, version, kind)
	client := cfg.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("ensure asset: request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ensure asset: download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("ensure asset: download %s: HTTP %d: %s", url, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	parent := filepath.Dir(destDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", fmt.Errorf("ensure asset: mkdir: %w", err)
	}
	tmp, err := os.MkdirTemp(parent, ".tmp-ensure-*")
	if err != nil {
		return "", fmt.Errorf("ensure asset: temp: %w", err)
	}
	defer os.RemoveAll(tmp)

	if err := extractTarGz(resp.Body, tmp); err != nil {
		return "", fmt.Errorf("ensure asset: extract: %w", err)
	}
	if !spaDirComplete(tmp) {
		return "", fmt.Errorf("ensure asset: extracted archive is not a complete SPA dist (need non-empty index.html and assets/)")
	}

	// Atomic replace.
	bak := destDir + ".bak-" + filepath.Base(tmp)
	_ = os.RemoveAll(bak)
	if _, err := os.Stat(destDir); err == nil {
		if err := os.Rename(destDir, bak); err != nil {
			return "", fmt.Errorf("ensure asset: backup: %w", err)
		}
	}
	if err := os.Rename(tmp, destDir); err != nil {
		_ = os.Rename(bak, destDir)
		return "", fmt.Errorf("ensure asset: commit: %w", err)
	}
	_ = os.RemoveAll(bak)
	// MkdirTemp was removed from tmp path by rename; prevent double-remove issues.
	// defer RemoveAll on renamed-away path is fine (no-op / ENOENT).

	return destDir, nil
}

func extractTarGz(r io.Reader, destDir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(filepath.ToSlash(hdr.Name), "./")
		if name == "" || name == "." {
			continue
		}
		// Prevent path traversal.
		if strings.Contains(name, "..") {
			return fmt.Errorf("illegal path in archive: %s", hdr.Name)
		}
		target := filepath.Join(destDir, filepath.FromSlash(name))
		if !strings.HasPrefix(target, destDir+string(os.PathSeparator)) && target != destDir {
			return fmt.Errorf("illegal path in archive: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			mode := hdr.FileInfo().Mode().Perm()
			if mode == 0 {
				mode = 0o644
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(f, tr)
			closeErr := f.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		default:
			// skip other types (symlinks, etc.)
		}
	}
}
