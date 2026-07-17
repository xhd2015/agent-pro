// Pack two release asset archives from fat frontend dist trees for hydrate downloads.
// Basenames from assets.AssetReleaseNames(version).
// Default: pack only. --upload opt-in wraps gh (create release if missing,
// gh release upload --clobber if exists). COPYFILE_DISABLE / exclude ._* .
//
// CLI (from module root):
//
//	go run ./script/github/release-assets [flags]
//
// Flags:
//
//	--out DIR         Output directory for archives; when omitted, create via
//	                  os.MkdirTemp("", "agent-pro-release-assets-*")
//	                  and print "out: <abs-path>" on stdout.
//	--version VER     Version string; normalized like AssetReleaseNames
//	                  (default: assets.ClientVersion() / pkgs/assets/VERSION.txt)
//	--upload          Opt-in GitHub upload via gh
//	-h, --help        Print usage; exit 0
//
// Pack sources (relative to module / process cwd = module root):
//
//	frontend-agent-run/dist  -> agent-run_v{ver}_frontend.tar.gz
//	frontend/dist            -> agent-pro_v{ver}_frontend.tar.gz
//
// Dist trees must be complete (non-empty index.html + non-empty assets/).
// Archive root = dist contents (index.html at archive root).
package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/assets"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Avoid macOS resource-fork AppleDouble noise in archives / copy trees.
	_ = os.Setenv("COPYFILE_DISABLE", "1")

	var (
		outDir  string
		version string
		upload  bool
		help    bool
	)

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			help = true
		case a == "--upload":
			upload = true
		case a == "--out":
			if i+1 >= len(args) {
				return fmt.Errorf("--out requires a directory argument")
			}
			i++
			outDir = args[i]
		case strings.HasPrefix(a, "--out="):
			outDir = strings.TrimPrefix(a, "--out=")
		case a == "--version":
			if i+1 >= len(args) {
				return fmt.Errorf("--version requires a value")
			}
			i++
			version = args[i]
		case strings.HasPrefix(a, "--version="):
			version = strings.TrimPrefix(a, "--version=")
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unrecognized flag %s (try --help)", a)
		default:
			return fmt.Errorf("unexpected argument %q (try --help)", a)
		}
	}

	if help {
		printHelp(os.Stdout)
		return nil
	}

	if version == "" {
		version = assets.ClientVersion()
	}
	version = assets.NormalizeVersion(version)
	if version == "" {
		return fmt.Errorf("--version is required (or set pkgs/assets/VERSION.txt)")
	}

	names := assets.AssetReleaseNames(version)
	if len(names) != 2 {
		return fmt.Errorf("AssetReleaseNames(%q) returned %d names want 2: %v", version, len(names), names)
	}

	if outDir == "" {
		tmp, err := os.MkdirTemp("", "agent-pro-release-assets-*")
		if err != nil {
			return fmt.Errorf("mkdir temp --out: %w", err)
		}
		outDir = tmp
	} else {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return fmt.Errorf("mkdir --out: %w", err)
		}
	}
	absOut, err := filepath.Abs(outDir)
	if err != nil {
		return fmt.Errorf("resolve --out absolute path: %w", err)
	}
	outDir = absOut

	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	// Source dirs aligned with AssetReleaseNames order:
	//   0 agent-run_*_frontend.tar.gz
	//   1 agent-pro_*_frontend.tar.gz
	sources := []string{
		filepath.Join(root, "frontend-agent-run", "dist"),
		filepath.Join(root, "frontend", "dist"),
	}

	var archives []string
	for i, src := range sources {
		if err := requireCompleteSPA(src); err != nil {
			return fmt.Errorf("pack source incomplete (%s): %w\n  Build fat dist first:\n    go run ./script/agent-run/build-frontend\n    go run ./script/build-frontend", src, err)
		}
		dest := filepath.Join(outDir, names[i])
		if err := writeTarGz(src, dest); err != nil {
			return fmt.Errorf("pack %s: %w", names[i], err)
		}
		info, err := os.Stat(dest)
		if err != nil {
			return fmt.Errorf("stat packed archive %s: %w", dest, err)
		}
		if info.Size() <= 0 {
			return fmt.Errorf("packed archive empty: %s", dest)
		}
		fmt.Printf("wrote %s (%d bytes)\n", dest, info.Size())
		archives = append(archives, dest)
	}

	fmt.Printf("out: %s\n", outDir)

	if !upload {
		return nil
	}
	return uploadWithGH(version, archives)
}

func printHelp(w io.Writer) {
	fmt.Fprint(w, `Usage: go run ./script/github/release-assets [flags]

Pack two release hydrate archives from fat frontend dist trees. Basenames match
assets.AssetReleaseNames(version).

Flags:
  --out DIR         Output directory for .tar.gz archives (default: temp dir
                    via MkdirTemp agent-pro-release-assets-*; prints out: path)
  --version VER     Version string (default: assets.ClientVersion / VERSION.txt)
  --upload          Opt-in: wrap gh to create the release tag if missing and
                    upload archives with --clobber
  -h, --help        Show this help

Sources (relative to module root / process cwd):
  frontend-agent-run/dist  -> agent-run_v{ver}_frontend.tar.gz
  frontend/dist            -> agent-pro_v{ver}_frontend.tar.gz

Requires complete SPA dist (non-empty index.html and assets/). Default is
pack-only (no network). When --out is omitted, packs into a temp dir and prints
out: <abs-path>. COPYFILE_DISABLE is set; AppleDouble ._* entries are excluded.
`)
}

// requireCompleteSPA checks that dir looks like a fat Vite dist.
func requireCompleteSPA(dir string) error {
	st, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("missing or unreadable: %w", err)
	}
	if !st.IsDir() {
		return fmt.Errorf("not a directory")
	}
	indexPath := filepath.Join(dir, "index.html")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("need non-empty index.html: %w", err)
	}
	if len(bytesTrimSpace(data)) == 0 {
		return fmt.Errorf("index.html is empty (thin/placeholder dist)")
	}
	if !hasNonEmptyAssetFile(dir) {
		return fmt.Errorf("need at least one non-empty file under assets/ (thin/placeholder dist)")
	}
	return nil
}

func hasNonEmptyAssetFile(dir string) bool {
	assetsDir := filepath.Join(dir, "assets")
	var found bool
	_ = filepath.WalkDir(assetsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		f, openErr := os.Open(path)
		if openErr != nil {
			return nil
		}
		defer f.Close()
		var buf [1]byte
		n, _ := f.Read(buf[:])
		if n > 0 {
			found = true
			return io.EOF
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

// writeTarGz packs the contents of srcDir into destPath as gzip-compressed tar.
// Paths inside the archive are relative to srcDir (index.html at archive root).
// Skips AppleDouble ._* files and .DS_Store.
func writeTarGz(srcDir, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		base := filepath.Base(path)
		if strings.HasPrefix(base, "._") || base == ".DS_Store" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		name := filepath.ToSlash(rel)

		if info.IsDir() {
			hdr, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			hdr.Name = name + "/"
			return tw.WriteHeader(hdr)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = name
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		rf, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tw, rf)
		closeErr := rf.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	if err != nil {
		_ = tw.Close()
		_ = gz.Close()
		_ = f.Close()
		_ = os.Remove(destPath)
		return err
	}
	if err := tw.Close(); err != nil {
		_ = gz.Close()
		_ = f.Close()
		_ = os.Remove(destPath)
		return err
	}
	if err := gz.Close(); err != nil {
		_ = f.Close()
		_ = os.Remove(destPath)
		return err
	}
	return f.Close()
}

// uploadWithGH creates the GitHub release for the version tag if missing, then
// uploads archives with clobber. Requires `gh` on PATH and authenticated repo.
func uploadWithGH(version string, archives []string) error {
	tag := assets.NormalizeVersion(version)
	if tag == "" {
		return fmt.Errorf("upload: empty version tag")
	}

	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("upload requires gh on PATH: %w", err)
	}

	view := exec.Command("gh", "release", "view", tag)
	if err := view.Run(); err != nil {
		create := exec.Command("gh", "release", "create", tag, "--title", tag, "--notes", "Hydrate asset archives for "+tag)
		create.Stdout = os.Stdout
		create.Stderr = os.Stderr
		if err := create.Run(); err != nil {
			return fmt.Errorf("gh release create %s: %w", tag, err)
		}
		fmt.Printf("created release %s\n", tag)
	}

	args := []string{"release", "upload", tag, "--clobber"}
	args = append(args, archives...)
	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh release upload %s: %w", tag, err)
	}
	fmt.Printf("uploaded %d archives to release %s\n", len(archives), tag)
	return nil
}
