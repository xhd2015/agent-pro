package run

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMirrorSessionsForWorkDir(t *testing.T) {
	grokHome := t.TempDir()
	workDir := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}

	abs, err := filepath.Abs(workDir)
	if err != nil {
		t.Fatal(err)
	}
	fromEnc := grokSessionEncoding(abs)
	toEnc := url.PathEscape(abs)
	if fromEnc == toEnc {
		t.Skip("workdir has no symlink components on this platform")
	}

	fromRoot := filepath.Join(grokHome, "sessions", fromEnc, "session-uuid")
	if err := os.MkdirAll(fromRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fromRoot, "events.jsonl"), []byte(`{"type":"turn_started","model_id":"mock-model"}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := mirrorSessionsForWorkDir(grokHome, workDir); err != nil {
		t.Fatal(err)
	}

	toEvents := filepath.Join(grokHome, "sessions", toEnc, "session-uuid", "events.jsonl")
	if _, err := os.Stat(toEvents); err != nil {
		t.Fatalf("mirrored events.jsonl missing at %s: %v", toEvents, err)
	}
}

func TestMirrorSessionsForWorkDirWaitsForSource(t *testing.T) {
	grokHome := t.TempDir()
	workDir := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(workDir)
	if err != nil {
		t.Fatal(err)
	}
	fromEnc := grokSessionEncoding(abs)
	toEnc := url.PathEscape(abs)
	if fromEnc == toEnc {
		t.Skip("symlink workdir required")
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		time.Sleep(100 * time.Millisecond)
		fromRoot := filepath.Join(grokHome, "sessions", fromEnc, "late-session")
		if err := os.MkdirAll(fromRoot, 0755); err != nil {
			t.Error(err)
			return
		}
		_ = os.WriteFile(filepath.Join(fromRoot, "events.jsonl"), []byte("line\n"), 0644)
	}()

	if err := MirrorSessionsForWorkDirWithRetry(grokHome, workDir, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	<-done

	toEvents := filepath.Join(grokHome, "sessions", toEnc, "late-session", "events.jsonl")
	if _, err := os.Stat(toEvents); err != nil {
		t.Fatalf("expected mirrored late session: %v", err)
	}
}