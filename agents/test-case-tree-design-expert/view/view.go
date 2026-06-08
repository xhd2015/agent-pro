package view

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/xhd2015/less-flags"
)

//go:embed tree_view.html
var treeViewHTML string

const viewUsage = `Usage: test-case-tree-design-expert view <dir>

Start a web server to browse the test case tree at <dir>.
`

func Run(args []string) error {
	args, err := lessflags.Help("-h,--help", viewUsage).Parse(args)
	if err != nil {
		return err
	}
	if len(args) < 1 {
		return fmt.Errorf("missing <dir>")
	}

	dir := args[0]
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory not found: %s", dir)
		}
		return fmt.Errorf("cannot stat directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("cannot resolve absolute path: %w", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/tree", func(w http.ResponseWriter, r *http.Request) {
		tree, err := buildTree(absDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, tree)
	})

	mux.HandleFunc("/api/file", func(w http.ResponseWriter, r *http.Request) {
		relPath := r.URL.Query().Get("path")
		if relPath == "" {
			http.Error(w, "missing path parameter", http.StatusBadRequest)
			return
		}
		fullPath := filepath.Join(absDir, relPath)
		if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(absDir)+string(filepath.Separator)) {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		data, err := os.ReadFile(fullPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, map[string]string{"content": string(data)})
	})

	mux.HandleFunc("/api/chain", func(w http.ResponseWriter, r *http.Request) {
		relPath := r.URL.Query().Get("path")
		if relPath == "" {
			http.Error(w, "missing path parameter", http.StatusBadRequest)
			return
		}
		parts := splitPath(relPath)
		var chain []chainEntry
		for i := 0; i <= len(parts); i++ {
			subPath := filepath.Join(append([]string{absDir}, parts[:i]...)...)
			setupPath := filepath.Join(subPath, "SETUP.md")
			data, err := os.ReadFile(setupPath)
			if err != nil {
				continue
			}
			source := "root"
			if i > 0 {
				source = parts[i-1]
			}
			if i == len(parts) {
				source = source + " (local)"
			}
			chain = append(chain, chainEntry{
				Source:  source,
				Content: string(data),
			})
		}
		writeJSON(w, chain)
	})

	mux.HandleFunc("/api/assert", func(w http.ResponseWriter, r *http.Request) {
		relPath := r.URL.Query().Get("path")
		if relPath == "" {
			http.Error(w, "missing path parameter", http.StatusBadRequest)
			return
		}
		fullPath := filepath.Join(absDir, relPath, "ASSERT.md")
		if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(absDir)+string(filepath.Separator)) {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		data, err := os.ReadFile(fullPath)
		if err != nil {
			writeJSON(w, map[string]string{"content": "no assert"})
			return
		}
		writeJSON(w, map[string]string{"content": string(data)})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(treeViewHTML))
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("cannot start server: %w", err)
	}

	addr := listener.Addr().(*net.TCPAddr)
	url := fmt.Sprintf("http://127.0.0.1:%d", addr.Port)
	fmt.Fprintf(os.Stderr, "Viewing test case tree: %s\n", absDir)
	fmt.Fprintf(os.Stderr, "Open %s in your browser\n", url)

	openBrowser(url)

	return http.Serve(listener, mux)
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
}

type treeNode struct {
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	Type      string      `json:"type"`
	Children  []*treeNode `json:"children,omitempty"`
}

type chainEntry struct {
	Source  string `json:"source"`
	Content string `json:"content"`
}

func buildTree(root string) (*treeNode, error) {
	return buildNode(root, root)
}

func buildNode(root, dir string) (*treeNode, error) {
	relPath, err := filepath.Rel(root, dir)
	if err != nil {
		return nil, err
	}
	if relPath == "." {
		relPath = ""
	}

	name := filepath.Base(dir)
	if relPath == "" {
		name = filepath.Base(dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	nodeType := "leaf"
	var children []*treeNode
	var subdirs []os.DirEntry

	for _, e := range entries {
		if e.IsDir() {
			subdirs = append(subdirs, e)
		}
	}

	if len(subdirs) > 0 {
		nodeType = "dir"
		sort.Slice(subdirs, func(i, j int) bool {
			return subdirs[i].Name() < subdirs[j].Name()
		})
		for _, e := range subdirs {
			child, err := buildNode(root, filepath.Join(dir, e.Name()))
			if err != nil {
				return nil, err
			}
			children = append(children, child)
		}
	}

	if relPath == "" {
		nodeType = "root"
	}

	return &treeNode{
		Name:     name,
		Path:     relPath,
		Type:     nodeType,
		Children: children,
	}, nil
}

func splitPath(p string) []string {
	if p == "" {
		return nil
	}
	return strings.Split(filepath.ToSlash(p), "/")
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(v)
}

func init() {
	_ = treeViewHTML
}
