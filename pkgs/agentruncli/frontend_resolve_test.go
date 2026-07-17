package agentruncli

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/xhd2015/agent-pro/pkgs/assets"
)

func TestResolveProductFrontend_preferEmbed(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv(assets.EnvAssetBaseURL, "")

	complete := fstest.MapFS{
		"dist/index.html":             &fstest.MapFile{Data: []byte("<html>embed</html>")},
		"dist/assets/index-abc.js":    &fstest.MapFile{Data: []byte("console.log(1)")},
	}
	// Also seed a complete cache so we can assert embed wins.
	cacheSrc := fstest.MapFS{
		"index.html":          &fstest.MapFile{Data: []byte("<html>cache</html>")},
		"assets/index-x.js":   &fstest.MapFile{Data: []byte("x")},
	}
	if err := assets.WriteAssetCache(assets.ProductAgentRun, assets.ClientVersion(), assets.KindFrontend, cacheSrc); err != nil {
		t.Fatal(err)
	}

	r := resolveProductFrontend(context.Background(), assets.ProductAgentRun, func() bool { return true }, complete, assets.EnsureConfig{})
	if r.Source != FrontendSourceEmbed {
		t.Fatalf("source = %q, want embed", r.Source)
	}
	if r.FS == nil {
		t.Fatal("expected FS")
	}
	data, err := fs.ReadFile(r.FS, "index.html")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "embed") {
		t.Fatalf("index = %q, want embed content", data)
	}
}

func TestResolveProductFrontend_preferCacheOverDownload(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv(assets.EnvAssetBaseURL, "http://127.0.0.1:1/should-not-hit")

	cacheSrc := fstest.MapFS{
		"index.html":        &fstest.MapFile{Data: []byte("<html>from-cache</html>")},
		"assets/index-y.js": &fstest.MapFile{Data: []byte("y")},
	}
	if err := assets.WriteAssetCache(assets.ProductAgentRun, assets.ClientVersion(), assets.KindFrontend, cacheSrc); err != nil {
		t.Fatal(err)
	}

	thin := fstest.MapFS{
		"dist/placeholder.txt": &fstest.MapFile{Data: []byte("thin\n")},
	}
	r := resolveProductFrontend(context.Background(), assets.ProductAgentRun, func() bool { return false }, thin, assets.EnsureConfig{})
	if r.Source != FrontendSourceCache {
		t.Fatalf("source = %q, want cache", r.Source)
	}
	data, err := fs.ReadFile(r.FS, "index.html")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "from-cache") {
		t.Fatalf("index = %q", data)
	}
}

func TestResolveProductFrontend_emptyBaseURL_A1(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv(assets.EnvAssetBaseURL, "")

	thin := fstest.MapFS{
		"dist/placeholder.txt": &fstest.MapFile{Data: []byte("thin\n")},
	}
	r := resolveProductFrontend(context.Background(), assets.ProductAgentRun, func() bool { return false }, thin, assets.EnsureConfig{})
	if r.Source != FrontendSourceA1 {
		t.Fatalf("source = %q, want a1", r.Source)
	}
	if r.FS != nil {
		t.Fatal("expected nil FS for A1")
	}
	if r.EnsureErr == nil {
		t.Fatal("expected EnsureErr when BaseURL empty")
	}
}

func TestA1IncompleteHTML_mentionsEnv(t *testing.T) {
	html := a1IncompleteHTML("agent-run")
	for _, want := range []string{
		"agent-run",
		assets.EnvAssetBaseURL,
		"assets ensure",
		"incomplete",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("A1 HTML missing %q:\n%s", want, html)
		}
	}
}

func TestResolveProductFrontend_cacheDirExistsOnDisk(t *testing.T) {
	// Sanity: CacheComplete after WriteAssetCache.
	root := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", root)
	src := fstest.MapFS{
		"index.html":        &fstest.MapFile{Data: []byte("<html>ok</html>")},
		"assets/index-z.js": &fstest.MapFile{Data: []byte("z")},
	}
	if err := assets.WriteAssetCache(assets.ProductAgentRun, "0.0.70", assets.KindFrontend, src); err != nil {
		t.Fatal(err)
	}
	dir, err := assets.AssetCacheDir(assets.ProductAgentRun, "0.0.70", assets.KindFrontend)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "index.html")); err != nil {
		t.Fatal(err)
	}
}
