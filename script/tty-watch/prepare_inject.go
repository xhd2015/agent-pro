package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

func registerPrepareInjectAPI(mux *http.ServeMux, mgr *ptywrap.Manager) {
	mux.HandleFunc("POST /api/terminal/sessions/{sessionID}/prepare-inject", func(w http.ResponseWriter, r *http.Request) {
		sessionID := strings.TrimSpace(r.PathValue("sessionID"))
		if sessionID == "" {
			http.Error(w, "missing session id", http.StatusBadRequest)
			return
		}
		if err := ensureSessionPTYInjectMode(mgr, sessionID); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func prepareSessionInjectMode(listenAddr, sessionID string) error {
	url := fmt.Sprintf("http://%s/api/terminal/sessions/%s/prepare-inject", listenAddr, sessionID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("prepare inject endpoint not found")
	}
	return fmt.Errorf("prepare inject failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
}