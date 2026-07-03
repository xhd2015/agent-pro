package ttyrunner

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// InjectInput writes input to a live session via server-side HTTP inject API.
func InjectInput(listenAddr, sessionID string, input []byte) error {
	url := fmt.Sprintf("http://%s/api/terminal/sessions/%s/input", listenAddr, sessionID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(input))
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
		return fmt.Errorf("inject endpoint not found")
	}
	return fmt.Errorf("inject failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
}