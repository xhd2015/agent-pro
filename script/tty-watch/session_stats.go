package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

var sessionStatsHTTPClient = &http.Client{Timeout: 2 * time.Second}

func fetchSessionClientCounts(listenAddr, sessionID string) (watch, attached int) {
	url := fmt.Sprintf("http://%s/api/terminal/sessions?page_size=100", listenAddr)
	resp, err := sessionStatsHTTPClient.Get(url)
	if err != nil {
		return 0, 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, 0
	}
	var body ptywrap.SessionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, 0
	}
	for _, session := range body.Sessions {
		if session.ID != sessionID {
			continue
		}
		watch = session.ObserverCount
		attached = session.AttacherCount
		if session.WriterConnected {
			attached++
		}
		return watch, attached
	}
	return 0, 0
}