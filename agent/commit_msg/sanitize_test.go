package commit_msg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitize_AntiPatternsFixtures(t *testing.T) {
	dir := filepath.Join("testdata", "anti_patterns")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".in") {
			names = append(names, strings.TrimSuffix(e.Name(), ".in"))
		}
	}
	if len(names) == 0 {
		t.Fatal("no fixtures found")
	}

	for _, name := range names {
		name := name
		t.Run(name, func(t *testing.T) {
			inBytes, err := os.ReadFile(filepath.Join(dir, name+".in"))
			if err != nil {
				t.Fatalf("read .in: %v", err)
			}
			in := string(inBytes)

			wantErrPath := filepath.Join(dir, name+".want_err")
			if _, err := os.Stat(wantErrPath); err == nil {
				_, sErr := SanitizeOrError(in)
				if sErr == nil {
					t.Fatalf("expected error for %s, got success", name)
				}
				wantSubBytes, _ := os.ReadFile(wantErrPath)
				wantSub := strings.TrimSpace(string(wantSubBytes))
				if wantSub != "" && !strings.Contains(strings.ToLower(sErr.Error()), strings.ToLower(wantSub)) {
					t.Logf("error %q does not contain %q (soft)", sErr.Error(), wantSub)
				}
				return
			}

			wantBytes, err := os.ReadFile(filepath.Join(dir, name+".want"))
			if err != nil {
				t.Fatalf("read .want: %v", err)
			}
			want := strings.TrimRight(string(wantBytes), "\n")

			msg, sErr := SanitizeOrError(in)
			if sErr != nil {
				t.Fatalf("unexpected error: %v", sErr)
			}
			got := msg.format()
			if got != want {
				t.Fatalf("format mismatch\n got: %q\nwant: %q", got, want)
			}
		})
	}
}
