package sessions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SessionFile is one regular file in a session directory.
type SessionFile struct {
	Name  string
	Size  int64
	Mtime time.Time
	Path  string
}

// ListSessionFiles returns the absolute session directory and a listing of
// regular files under it. Unknown session → error containing
// "grok session not found" and the id.
func ListSessionFiles(grokHome, sessionID string) (dir string, files []SessionFile, err error) {
	session, err := Find(grokHome, sessionID)
	if err != nil {
		return "", nil, err
	}
	dir = filepath.Dir(session.Path)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return dir, nil, fmt.Errorf("readdir session: %w", err)
	}

	files = make([]SessionFile, 0, len(entries))
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		info, err := ent.Info()
		if err != nil {
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		name := ent.Name()
		path := filepath.Join(dir, name)
		files = append(files, SessionFile{
			Name:  name,
			Size:  info.Size(),
			Mtime: info.ModTime(),
			Path:  path,
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})
	return dir, files, nil
}

// FormatSessionFilesTable prints a human table with NAME, SIZE, MTIME columns.
func FormatSessionFilesTable(files []SessionFile) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-24s  %10s  %s\n", "NAME", "SIZE", "MTIME")
	for _, f := range files {
		fmt.Fprintf(
			&b,
			"%-24s  %10d  %s\n",
			f.Name,
			f.Size,
			f.Mtime.UTC().Format(time.RFC3339),
		)
	}
	return strings.TrimRight(b.String(), "\n")
}

// FormatSessionFilesJSON renders files as a JSON array (no ANSI).
func FormatSessionFilesJSON(files []SessionFile) (string, error) {
	type fileJSON struct {
		Name  string `json:"name"`
		Size  int64  `json:"size"`
		Mtime string `json:"mtime"`
		Path  string `json:"path,omitempty"`
	}
	out := make([]fileJSON, 0, len(files))
	for _, f := range files {
		out = append(out, fileJSON{
			Name:  f.Name,
			Size:  f.Size,
			Mtime: f.Mtime.UTC().Format(time.RFC3339),
			Path:  f.Path,
		})
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
