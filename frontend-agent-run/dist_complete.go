package frontend

import (
	"io"
	"io/fs"
	"strings"
)

// DistComplete reports whether the embedded SPA under DistFS looks like a real
// (fat) build rather than a thin placeholder-only embed.
//
// Complete when dist contains a non-empty index.html and at least one non-empty
// file under assets/ (or similar real build artifact path).
func DistComplete() bool {
	return distComplete(DistFS)
}

func distComplete(root fs.FS) bool {
	dist, err := fs.Sub(root, "dist")
	if err != nil {
		return false
	}
	return spaDistComplete(dist)
}

// spaDistComplete inspects a filesystem whose root is the SPA dist directory
// (i.e. index.html lives at "index.html", not "dist/index.html").
func spaDistComplete(dist fs.FS) bool {
	index, err := fs.ReadFile(dist, "index.html")
	if err != nil || len(bytesTrimSpace(index)) == 0 {
		return false
	}
	return hasNonEmptyAsset(dist)
}

func hasNonEmptyAsset(dist fs.FS) bool {
	var found bool
	_ = fs.WalkDir(dist, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Missing assets/ → incomplete.
			return err
		}
		if d.IsDir() || path == "assets" {
			return nil
		}
		if !strings.HasPrefix(path, "assets/") {
			return nil
		}
		f, err := dist.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		var buf [1]byte
		n, _ := f.Read(buf[:])
		if n > 0 {
			found = true
			return io.EOF // stop walk
		}
		return nil
	})
	return found
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
