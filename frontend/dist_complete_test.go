package frontend

import (
	"testing"
	"testing/fstest"
)

func TestDistComplete_placeholderOnly(t *testing.T) {
	fsys := fstest.MapFS{
		"dist/placeholder.txt": &fstest.MapFile{
			Data: []byte("thin embed placeholder\n"),
		},
	}
	if distComplete(fsys) {
		t.Fatal("expected placeholder-only dist to be incomplete")
	}
}

func TestDistComplete_indexAndAsset(t *testing.T) {
	fsys := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{
			Data: []byte("<!doctype html><title>app</title>\n"),
		},
		"dist/assets/x.js": &fstest.MapFile{
			Data: []byte("console.log(1)\n"),
		},
	}
	if !distComplete(fsys) {
		t.Fatal("expected index.html + assets/x.js to be complete")
	}
}

func TestDistComplete_emptyIndex(t *testing.T) {
	fsys := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{
			Data: []byte("   \n"),
		},
		"dist/assets/x.js": &fstest.MapFile{
			Data: []byte("ok"),
		},
	}
	if distComplete(fsys) {
		t.Fatal("expected empty index.html to be incomplete")
	}
}

func TestDistComplete_indexWithoutAssets(t *testing.T) {
	fsys := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{
			Data: []byte("<!doctype html>\n"),
		},
		"dist/placeholder.txt": &fstest.MapFile{
			Data: []byte("only placeholder\n"),
		},
	}
	if distComplete(fsys) {
		t.Fatal("expected index without assets to be incomplete")
	}
}

func TestDistComplete_liveEmbed(t *testing.T) {
	// Live DistFS may be fat (local build present) or thin (placeholder only).
	// Just ensure DistComplete() does not panic and matches distComplete(DistFS).
	got := DistComplete()
	want := distComplete(DistFS)
	if got != want {
		t.Fatalf("DistComplete()=%v, distComplete(DistFS)=%v", got, want)
	}
}
