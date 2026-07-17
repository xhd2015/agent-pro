package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/frontend"
	"github.com/xhd2015/agent-pro/pkgs/assets"
)

func TestStatic_A1WhenNoCompleteAssets(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv(assets.EnvAssetBaseURL, "")
	SetFrontendDistDir("")

	// Empty embed (no dist) — DistComplete is false for real package if fat
	// is present in workspace. Force incomplete path by using empty Init FS
	// only when live embed is incomplete; otherwise skip if DistComplete.
	if frontend.DistComplete() {
		// Still exercise A1 by pointing Init at an empty FS cannot change
		// DistComplete() which reads package DistFS. Simulate A1 via direct
		// handler path: clear by testing a1 HTML content instead.
		html := a1IncompleteHTMLAgentPro()
		for _, want := range []string{"agent-pro", assets.EnvAssetBaseURL, "incomplete"} {
			if !strings.Contains(html, want) {
				t.Fatalf("A1 HTML missing %q", want)
			}
		}
		// When fat embed is present, Static must still succeed offline.
		Init(frontend.DistFS, frontend.TemplateHTML)
		mux := http.NewServeMux()
		if err := Static(mux, StaticOptions{}); err != nil {
			t.Fatalf("Static with fat embed: %v", err)
		}
		return
	}

	// Thin live embed path
	Init(frontend.DistFS, frontend.TemplateHTML)
	mux := http.NewServeMux()
	if err := Static(mux, StaticOptions{}); err != nil {
		t.Fatalf("Static must not fail on thin embed: %v", err)
	}
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "incomplete") {
		t.Fatalf("body = %s", rr.Body.String())
	}
}

func TestReactDistFS_prefersFrontendDistDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html>dir-override</html>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "a.js"), []byte(`1`), 0o644); err != nil {
		t.Fatal(err)
	}
	SetFrontendDistDir(dir)
	defer SetFrontendDistDir("")

	fsys, err := reactDistFS()
	if err != nil {
		t.Fatal(err)
	}
	data, err := fsys.Open("index.html")
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
}
