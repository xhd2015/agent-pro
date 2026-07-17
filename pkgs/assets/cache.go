package assets

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// AssetCacheRoot returns the root directory for downloaded SPA assets:
//   $XDG_CACHE_HOME/agent-pro/asset-cache
// or, when XDG_CACHE_HOME is unset:
//   ~/.cache/agent-pro/asset-cache
// (via os.UserCacheDir).
func AssetCacheRoot() (string, error) {
	if xdg := strings.TrimSpace(os.Getenv("XDG_CACHE_HOME")); xdg != "" {
		return filepath.Join(xdg, "agent-pro", "asset-cache"), nil
	}
	base, err := os.UserCacheDir()
	if err != nil {
		// Fallback when UserCacheDir fails: home/.cache
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return "", fmt.Errorf("asset cache root: %w", err)
		}
		base = filepath.Join(home, ".cache")
	}
	return filepath.Join(base, "agent-pro", "asset-cache"), nil
}

// AssetCacheDir returns the cache directory for a specific product/version/kind.
// Layout: {AssetCacheRoot}/{product}/{version}/{kind}
// Version is normalized to a leading "v".
func AssetCacheDir(product, version, kind string) (string, error) {
	root, err := AssetCacheRoot()
	if err != nil {
		return "", err
	}
	v := NormalizeVersion(version)
	if product == "" || v == "" || kind == "" {
		return "", fmt.Errorf("asset cache dir: product, version, and kind are required")
	}
	return filepath.Join(root, product, v, kind), nil
}

// CacheComplete reports whether the on-disk SPA cache for product/version/kind
// looks like a real build: non-empty index.html and at least one non-empty file
// under assets/ (same idea as frontend DistComplete / spaDistComplete).
func CacheComplete(product, version, kind string) bool {
	dir, err := AssetCacheDir(product, version, kind)
	if err != nil {
		return false
	}
	return spaDirComplete(dir)
}

// spaDirComplete inspects a directory whose root is the SPA dist tree
// (index.html at root).
func spaDirComplete(dir string) bool {
	indexPath := filepath.Join(dir, "index.html")
	data, err := os.ReadFile(indexPath)
	if err != nil || len(bytesTrimSpace(data)) == 0 {
		return false
	}
	return hasNonEmptyAssetDir(dir)
}

func hasNonEmptyAssetDir(dir string) bool {
	assetsDir := filepath.Join(dir, "assets")
	var found bool
	_ = filepath.WalkDir(assetsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(assetsDir, path)
		if relErr != nil || rel == "." {
			return nil
		}
		f, openErr := os.Open(path)
		if openErr != nil {
			return nil
		}
		defer f.Close()
		var buf [1]byte
		n, _ := f.Read(buf[:])
		if n > 0 {
			found = true
			return io.EOF
		}
		return nil
	})
	return found
}

// WriteAssetCache writes the contents of src (root = dist tree with index.html)
// into the cache directory for product/version/kind using atomic temp+rename.
func WriteAssetCache(product, version, kind string, src fs.FS) error {
	dest, err := AssetCacheDir(product, version, kind)
	if err != nil {
		return err
	}
	return writeFSAtomic(dest, src)
}

// writeFSAtomic copies src into destDir via a sibling temp directory then renames.
func writeFSAtomic(destDir string, src fs.FS) error {
	parent := filepath.Dir(destDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("asset cache mkdir: %w", err)
	}

	tmp, err := os.MkdirTemp(parent, ".tmp-asset-*")
	if err != nil {
		return fmt.Errorf("asset cache temp: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(tmp)
		}
	}()

	if err := copyFS(tmp, src); err != nil {
		return err
	}

	// Replace existing dest: rename dest aside, rename tmp into place, remove old.
	bak := destDir + ".bak-" + filepath.Base(tmp)
	_ = os.RemoveAll(bak)
	if _, err := os.Stat(destDir); err == nil {
		if err := os.Rename(destDir, bak); err != nil {
			return fmt.Errorf("asset cache backup: %w", err)
		}
	}
	if err := os.Rename(tmp, destDir); err != nil {
		// try restore
		_ = os.Rename(bak, destDir)
		return fmt.Errorf("asset cache commit: %w", err)
	}
	cleanup = false
	_ = os.RemoveAll(bak)
	return nil
}

func copyFS(destDir string, src fs.FS) error {
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		// Skip hidden macOS junk if present
		base := filepath.Base(path)
		if base == ".DS_Store" {
			return nil
		}
		target := filepath.Join(destDir, filepath.FromSlash(path))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := src.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}

func bytesTrimSpace(b []byte) []byte {
	i, j := 0, len(b)
	for i < j && (b[i] == ' ' || b[i] == '\t' || b[i] == '\n' || b[i] == '\r') {
		i++
	}
	for j > i && (b[j-1] == ' ' || b[j-1] == '\t' || b[j-1] == '\n' || b[j-1] == '\r') {
		j--
	}
	return b[i:j]
}
