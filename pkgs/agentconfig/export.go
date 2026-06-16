package agentconfig

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Export(agent string, homeDir string, zipPath string) error {
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		return fmt.Errorf("create zip dir: %w", err)
	}
	fw, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("create zip: %w", err)
	}
	defer fw.Close()
	zw := zip.NewWriter(fw)
	defer zw.Close()

	switch agent {
	case "opencode":
		return exportOpencode(zw, homeDir)
	case "pi":
		return exportPi(zw, homeDir)
	case "crush":
		return exportCrush(zw, homeDir)
	default:
		return fmt.Errorf("unknown agent: %s", agent)
	}
}

func exportOpencode(zw *zip.Writer, homeDir string) error {
	// *.json files from ~/.local/share/opencode/
	dataDir := filepath.Join(homeDir, ".local", "share", "opencode")
	if entries, err := os.ReadDir(dataDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".json") {
				continue
			}
			srcPath := filepath.Join(dataDir, name)
			zipEntry := "opencode/" + name
			if err := addFileToZip(zw, srcPath, zipEntry); err != nil {
				return err
			}
		}
	}

	// opencode.jsonc from ~/.config/opencode/
	cfgDir := filepath.Join(homeDir, ".config", "opencode")
	cfgPath := filepath.Join(cfgDir, "opencode.jsonc")
	if _, err := os.Stat(cfgPath); err == nil {
		if err := addFileToZip(zw, cfgPath, "opencode/opencode.jsonc"); err != nil {
			return err
		}
	}

	// plugins from ~/.config/opencode/plugins/
	pluginsDir := filepath.Join(cfgDir, "plugins")
	if pluginEntries, err := os.ReadDir(pluginsDir); err == nil {
		for _, entry := range pluginEntries {
			if entry.IsDir() {
				continue
			}
			srcPath := filepath.Join(pluginsDir, entry.Name())
			zipEntry := "opencode/plugins/" + entry.Name()
			if err := addFileToZip(zw, srcPath, zipEntry); err != nil {
				return err
			}
		}
	}

	// skills from ~/.config/opencode/skills/ (recursive)
	skillsDir := filepath.Join(cfgDir, "skills")
	if err := filepath.Walk(skillsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(skillsDir, path)
		if err != nil {
			return nil
		}
		zipEntry := "opencode/skills/" + filepath.ToSlash(rel)
		return addFileToZip(zw, path, zipEntry)
	}); err != nil {
		return nil
	}

	return nil
}

func exportPi(zw *zip.Writer, homeDir string) error {
	agentDir := filepath.Join(homeDir, ".pi", "agent")
	files := []string{"auth.json", "settings.json", "models.json"}
	for _, f := range files {
		srcPath := filepath.Join(agentDir, f)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}
		zipEntry := "pi/" + f
		if err := addFileToZip(zw, srcPath, zipEntry); err != nil {
			return err
		}
	}
	return nil
}

func exportCrush(zw *zip.Writer, homeDir string) error {
	entries := []struct {
		source string
		zip    string
	}{
		{filepath.Join(homeDir, ".config", "crush", "crush.json"), "crush/config/crush.json"},
		{filepath.Join(homeDir, ".local", "share", "crush", "crush.json"), "crush/data/crush.json"},
	}
	for _, e := range entries {
		if _, err := os.Stat(e.source); os.IsNotExist(err) {
			continue
		}
		if err := addFileToZip(zw, e.source, e.zip); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zw *zip.Writer, srcPath string, zipPath string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", srcPath, err)
	}
	fw, err := zw.Create(zipPath)
	if err != nil {
		return fmt.Errorf("create zip entry %s: %w", zipPath, err)
	}
	if _, err := fw.Write(data); err != nil {
		return fmt.Errorf("write zip entry %s: %w", zipPath, err)
	}
	return nil
}
