package sessions

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/klauspost/compress/zstd"
)

// zstdStore writes the same body as one zstd frame per session.
func zstdStore(t *testing.T, sessions map[string]string) string {
	t.Helper()
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderCRC(true))
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()
	root := t.TempDir()
	for dir, body := range sessions {
		path := filepath.Join(root, "project", dir, "session.v3.jsonl.zstd")
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, enc.EncodeAll([]byte(body), nil), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestPooledStreamDecoderReusesAcrossArtifacts checks that the pooled decoder
// returns the same identity for every artifact, including on repeated walks,
// so a reused decoder cannot leak state into the next file.
func TestPooledStreamDecoderReusesAcrossArtifacts(t *testing.T) {
	root := zstdStore(t, map[string]string{
		"a": headerAt("id-a", 3) + row("turn/start", 0, `{"turn":1}`),
		"b": headerAt("id-b", 3) + row("session/title", 0, `{"title":"second"}`) + row("turn/start", 1, `{"turn":1}`),
	})
	store := &Store{Root: root}
	arts, _, err := store.artifacts()
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 2 {
		t.Fatalf("artifacts=%d", len(arts))
	}
	for pass := 0; pass < 3; pass++ {
		for _, a := range arts {
			id, err := artifactID(a)
			if err != nil {
				t.Fatalf("pass %d %s: %v", pass, a.path, err)
			}
			if want := "id-" + filepath.Base(filepath.Dir(a.path)); id != want {
				t.Fatalf("pass %d id=%q want %q", pass, id, want)
			}
			sum, err := recentSummary(a)
			if err != nil {
				t.Fatalf("pass %d %s: %v", pass, a.path, err)
			}
			if sum.ID != id {
				t.Fatalf("pass %d summary id=%q", pass, sum.ID)
			}
		}
	}
	if got, err := recentSummary(arts[0]); err != nil || got.ID != "id-a" {
		t.Fatalf("%+v %v", got, err)
	}
}

// TestPooledStreamDecoderAvoidsPerCallWindow documents the win: decoding one
// small artifact repeatedly must not allocate the decoder window each time.
// A fresh decoder costs about 3MB per call, so the bound sits well below it.
func TestPooledStreamDecoderAvoidsPerCallWindow(t *testing.T) {
	root := zstdStore(t, map[string]string{"a": headerAt("a", 3) + row("turn/start", 0, `{"turn":1}`)})
	arts, _, err := (&Store{Root: root}).artifacts()
	if err != nil || len(arts) != 1 {
		t.Fatalf("%v %v", arts, err)
	}
	if _, err := artifactID(arts[0]); err != nil {
		t.Fatal(err)
	}
	const runs = 20
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := 0; i < runs; i++ {
		if _, err := artifactID(arts[0]); err != nil {
			t.Fatal(err)
		}
	}
	runtime.ReadMemStats(&after)
	perCall := (after.TotalAlloc - before.TotalAlloc) / runs
	if perCall > 1<<20 {
		t.Fatalf("per call allocation %d bytes, want under 1MB (decoder window not pooled)", perCall)
	}
}
