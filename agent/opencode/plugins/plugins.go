package plugins

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Info struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func Install(destDir, pluginPath string) (string, error) {
	src, err := os.Open(pluginPath)
	if err != nil {
		return "", fmt.Errorf("open source plugin: %w", err)
	}
	defer src.Close()

	baseName := filepath.Base(pluginPath)
	ext := filepath.Ext(baseName)
	if ext != ".ts" && ext != ".js" {
		return "", fmt.Errorf("plugin file must be .ts or .js, got: %s", ext)
	}

	pluginsDir := filepath.Join(destDir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		return "", fmt.Errorf("create plugins dir: %w", err)
	}

	dstPath := filepath.Join(pluginsDir, baseName)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("create destination plugin: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("copy plugin file: %w", err)
	}

	return dstPath, nil
}

func InstallToDir(opencodeDir, pluginPath string) (string, error) {
	return Install(opencodeDir, pluginPath)
}

func List(dir string) ([]Info, error) {
	patterns := []string{
		filepath.Join(dir, "plugin", "*.ts"),
		filepath.Join(dir, "plugin", "*.js"),
		filepath.Join(dir, "plugins", "*.ts"),
		filepath.Join(dir, "plugins", "*.js"),
	}

	var result []Info
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("glob %s: %w", pattern, err)
		}
		for _, m := range matches {
			name := filepath.Base(m)
			nameWithoutExt := strings.TrimSuffix(name, filepath.Ext(name))
			result = append(result, Info{Name: nameWithoutExt, Path: m})
		}
	}
	return result, nil
}
