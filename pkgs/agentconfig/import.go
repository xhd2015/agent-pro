package agentconfig

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Import(homeDir string, zipPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		zipPath := filepath.ToSlash(f.Name)
		dest := resolveDest(zipPath, homeDir)
		if dest == nil {
			continue
		}
		content, err := readZipFile(f)
		if err != nil {
			return fmt.Errorf("read %s from zip: %w", zipPath, err)
		}
		if err := os.MkdirAll(filepath.Dir(dest.destPath), 0755); err != nil {
			return fmt.Errorf("create dir for %s: %w", dest.destPath, err)
		}
		if err := os.WriteFile(dest.destPath, content, dest.perm); err != nil {
			return fmt.Errorf("write %s: %w", dest.destPath, err)
		}
	}
	return nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
