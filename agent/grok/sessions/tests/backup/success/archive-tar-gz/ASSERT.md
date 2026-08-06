## Expected

- Backup succeeds.
- `Result.Dir` exists (temp dir kept) with `manifest.json`.
- `Result.ArchivePath` equals requested path; file exists and is non-empty.
- Archive is a readable `.tar.gz` containing at least `manifest.json`.

## Errors

- None.

```go
import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	dir := assertSuccessfulBackup(t, req, resp)

	if resp.Result.ArchivePath == "" {
		t.Fatal("Result.ArchivePath empty")
	}
	if filepath.Clean(resp.Result.ArchivePath) != filepath.Clean(req.ArchivePath) {
		t.Fatalf("ArchivePath = %q, want %q", resp.Result.ArchivePath, req.ArchivePath)
	}
	assertFileExists(t, req.ArchivePath)
	st, err := os.Stat(req.ArchivePath)
	if err != nil || st.Size() == 0 {
		t.Fatalf("archive missing or empty: %v size=%v", err, st)
	}

	// Dir kept after archive.
	assertDirExists(t, dir)
	assertFileExists(t, filepath.Join(dir, "manifest.json"))

	names := tarGzNames(t, req.ArchivePath)
	foundManifest := false
	for _, n := range names {
		if n == "manifest.json" || strings.HasSuffix(n, "/manifest.json") || filepath.Base(n) == "manifest.json" {
			foundManifest = true
			break
		}
	}
	if !foundManifest {
		t.Fatalf("archive missing manifest.json; entries=%v", names)
	}
}

func tarGzNames(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("gzip: %v", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	var names []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar next: %v", err)
		}
		names = append(names, hdr.Name)
	}
	return names
}
```
