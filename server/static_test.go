package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/frontend"
)

// TestStaticServesIndexFromSharedFrontendEmbed checks the live frontend.DistFS
// embed. Fat builds (DistComplete) must serve the real SPA shell; thin CI
// checkouts only ship dist/placeholder.txt, so we assert that marker instead
// of requiring #root (which would demand a full Vite embed in every go test).
func TestStaticServesIndexFromSharedFrontendEmbed(t *testing.T) {
	Init(frontend.DistFS, frontend.TemplateHTML)

	mux := http.NewServeMux()
	if err := Static(mux, StaticOptions{}); err != nil {
		t.Fatalf("static setup: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()

	if !frontend.DistComplete() {
		// Thin embed: no SPA index in DistFS. Placeholder keeps //go:embed valid.
		data, err := fs.ReadFile(frontend.DistFS, "dist/placeholder.txt")
		if err != nil {
			t.Fatalf("thin embed missing dist/placeholder.txt: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("thin embed dist/placeholder.txt is empty")
		}
		return
	}

	if !strings.Contains(body, `<div id="root">`) || !strings.Contains(body, "window.__KOOL_ROUTE_PREFIX__") {
		t.Fatalf("index HTML missing app shell/runtime prefix: %s", body)
	}
}

func TestStaticServesFrontendDistDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><head></head><body>local</body></html>`), 0o644); err != nil {
		t.Fatal(err)
	}
	SetFrontendDistDir(dir)
	defer SetFrontendDistDir("")

	mux := http.NewServeMux()
	if err := Static(mux, StaticOptions{}); err != nil {
		t.Fatalf("static: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "local") {
		t.Fatalf("body = %q, want local index", rec.Body.String())
	}
}
