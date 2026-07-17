package assets

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"testing/fstest"
)

func withTestCacheRoot(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", root)
}

func makeSPATarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(tw, content); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestCacheComplete_and_WriteAssetCache(t *testing.T) {
	withTestCacheRoot(t)

	if CacheComplete(ProductAgentPro, "0.0.70", KindFrontend) {
		t.Fatal("expected incomplete empty cache")
	}

	src := fstest.MapFS{
		"index.html":              &fstest.MapFile{Data: []byte("<html></html>")},
		"assets/index-abc123.js":  &fstest.MapFile{Data: []byte("console.log(1)")},
	}
	if err := WriteAssetCache(ProductAgentPro, "0.0.70", KindFrontend, src); err != nil {
		t.Fatal(err)
	}
	if !CacheComplete(ProductAgentPro, "0.0.70", KindFrontend) {
		t.Fatal("expected complete after WriteAssetCache")
	}
	dir, err := AssetCacheDir(ProductAgentPro, "0.0.70", KindFrontend)
	if err != nil {
		t.Fatal(err)
	}
	// version normalized in path
	if filepath.Base(filepath.Dir(dir)) != "v0.0.70" {
		t.Fatalf("cache path version segment: %s", dir)
	}
}

func TestEnsureAsset_completeCacheNoDownload(t *testing.T) {
	withTestCacheRoot(t)

	src := fstest.MapFS{
		"index.html":             &fstest.MapFile{Data: []byte("<html>ok</html>")},
		"assets/app.js":          &fstest.MapFile{Data: []byte("1")},
	}
	if err := WriteAssetCache(ProductAgentRun, "v0.0.1", KindFrontend, src); err != nil {
		t.Fatal(err)
	}

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Error(w, "should not be called", 500)
	}))
	defer srv.Close()

	dir, err := EnsureAsset(context.Background(), ProductAgentRun, "0.0.1", KindFrontend, EnsureConfig{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 0 {
		t.Fatalf("expected no HTTP hits, got %d", hits.Load())
	}
	if !spaDirComplete(dir) {
		t.Fatalf("dir not complete: %s", dir)
	}
}

func TestEnsureAsset_incompleteDownloads(t *testing.T) {
	withTestCacheRoot(t)

	archive := makeSPATarGz(t, map[string]string{
		"index.html":    "<!doctype html><html></html>",
		"assets/main.js": "export default 1",
	})

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		wantPath := "/v0.0.2/agent-pro_v0.0.2_frontend.tar.gz"
		if r.URL.Path != wantPath {
			http.Error(w, "bad path: "+r.URL.Path, 404)
			return
		}
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	if CacheComplete(ProductAgentPro, "0.0.2", KindFrontend) {
		t.Fatal("should start incomplete")
	}

	dir, err := EnsureAsset(context.Background(), ProductAgentPro, "0.0.2", KindFrontend, EnsureConfig{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits=%d, want 1", hits.Load())
	}
	if !spaDirComplete(dir) {
		t.Fatal("expected complete after download")
	}
	if !CacheComplete(ProductAgentPro, "v0.0.2", KindFrontend) {
		t.Fatal("CacheComplete false after ensure")
	}

	// Second call should not hit network.
	_, err = EnsureAsset(context.Background(), ProductAgentPro, "0.0.2", KindFrontend, EnsureConfig{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 {
		t.Fatalf("second ensure should not download: hits=%d", hits.Load())
	}
}

func TestEnsureAsset_missingBaseURL(t *testing.T) {
	withTestCacheRoot(t)
	t.Setenv(EnvAssetBaseURL, "")

	_, err := EnsureAsset(context.Background(), ProductAgentPro, "0.0.3", KindFrontend, EnsureConfig{})
	if err == nil {
		t.Fatal("expected error for missing BaseURL")
	}
	if !containsAll(err.Error(), "BaseURL", EnvAssetBaseURL) {
		t.Fatalf("error should mention BaseURL and env: %v", err)
	}
}

func TestEnsureAsset_baseURLFromEnv(t *testing.T) {
	withTestCacheRoot(t)

	archive := makeSPATarGz(t, map[string]string{
		"index.html":    "<html>x</html>",
		"assets/x.css":  "body{}",
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	t.Setenv(EnvAssetBaseURL, srv.URL)
	dir, err := EnsureAsset(context.Background(), ProductAgentRun, "1.2.3", KindFrontend, EnsureConfig{
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !spaDirComplete(dir) {
		t.Fatal("incomplete")
	}
}

func TestAssetCacheRoot_XDG(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", root)
	got, err := AssetCacheRoot()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "agent-pro", "asset-cache")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWriteAssetCache_incompleteNotComplete(t *testing.T) {
	withTestCacheRoot(t)
	src := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html></html>")},
		// no assets/
	}
	if err := WriteAssetCache(ProductAgentPro, "9.9.9", KindFrontend, src); err != nil {
		t.Fatal(err)
	}
	if CacheComplete(ProductAgentPro, "9.9.9", KindFrontend) {
		t.Fatal("index-only should be incomplete")
	}
}

func TestSpaDirComplete_emptyIndex(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("  \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "a.js"), []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if spaDirComplete(dir) {
		t.Fatal("empty index should be incomplete")
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !bytes.Contains([]byte(s), []byte(p)) {
			return false
		}
	}
	return true
}
