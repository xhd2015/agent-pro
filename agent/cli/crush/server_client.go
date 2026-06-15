package crush

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

type CrushServerClient struct {
	host           string
	httpClient     *http.Client
	clientID       string
	socketPath     string
	workspaceCache map[string]string
	mu             sync.Mutex
	serverStarted  bool
}

func getSocketPath() string {
	uid := strconv.Itoa(os.Getuid())
	return filepath.Join(os.TempDir(), "crush-"+uid+".sock")
}

func getDefaultHost() string {
	return "unix://" + getSocketPath()
}

func NewCrushServerClient() (*CrushServerClient, error) {
	socketPath := getSocketPath()
	client := &CrushServerClient{
		host:           "http://localhost",
		socketPath:     socketPath,
		clientID:       uuid.New().String(),
		workspaceCache: make(map[string]string),
	}
	client.httpClient = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
		Timeout: 30 * time.Second,
	}
	return client, nil
}

func (c *CrushServerClient) EnsureServer(ctx context.Context) error {
	healthy, err := c.probeHealth(ctx)
	if err == nil && healthy {
		c.mu.Lock()
		c.serverStarted = true
		c.mu.Unlock()
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := os.Stat(c.socketPath); err == nil {
		os.Remove(c.socketPath)
	}

	cmd := exec.Command("crush", "server")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start crush server: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("timed out waiting for crush server to become healthy")
		default:
			healthy, err := c.probeHealth(ctx)
			if err == nil && healthy {
				c.serverStarted = true
				return nil
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func (c *CrushServerClient) ServerStarted() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.serverStarted
}

func (c *CrushServerClient) probeHealth(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.host+"/v1/health", nil)
	if err != nil {
		return false, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func (c *CrushServerClient) HealthCheck(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.host+"/v1/health", nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func (c *CrushServerClient) CreateWorkspace(ctx context.Context, cwd string) (string, error) {
	c.mu.Lock()
	if id, ok := c.workspaceCache[cwd]; ok {
		c.mu.Unlock()
		return id, nil
	}
	c.mu.Unlock()

	body := map[string]string{
		"path":      cwd,
		"client_id": c.clientID,
	}
	bodyData, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", c.host+"/v1/workspaces", bytes.NewReader(bodyData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create workspace: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Message string `json:"message"`
		}
		json.Unmarshal(bodyBytes, &errBody)
		if errBody.Message != "" {
			return "", fmt.Errorf("create workspace: %s, body=%s", errBody.Message, string(bodyBytes))
		}
		return "", fmt.Errorf("create workspace: status %d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	var createResult struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(bodyBytes, &createResult); err != nil {
		return "", fmt.Errorf("decode workspace response: %w, body=%s", err, string(bodyBytes))
	}
	if createResult.ID == "" {
		return "", fmt.Errorf("create workspace: returned empty id, body=%s", string(bodyBytes))
	}

	c.mu.Lock()
	c.workspaceCache[cwd] = createResult.ID
	c.mu.Unlock()

	return createResult.ID, nil
}

func (c *CrushServerClient) SubscribeEvents(ctx context.Context, workspaceID string) (<-chan string, error) {
	subscribeClientID := uuid.New().String()
	req, err := http.NewRequestWithContext(ctx, "GET", c.host+"/v1/workspaces/"+workspaceID+"/events?client_id="+subscribeClientID, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("subscribe events: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("subscribe events: status %d", resp.StatusCode)
	}

	ch := make(chan string, 100)
	go func() {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			ch <- line
		}
		close(ch)
	}()
	return ch, nil
}

func (c *CrushServerClient) CreateSession(ctx context.Context, workspaceID string, title string) (string, error) {
	body := map[string]string{"title": title}
	bodyData, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", c.host+"/v1/workspaces/"+workspaceID+"/sessions", bytes.NewReader(bodyData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("create session: status %d", resp.StatusCode)
	}
	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode session response: %w", err)
	}
	return result.ID, nil
}

func (c *CrushServerClient) InitAgent(ctx context.Context, workspaceID string) error {
	url := c.host + "/v1/workspaces/" + workspaceID + "/agent/init"
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("init agent: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&errBody)
		return fmt.Errorf("init agent: status %d, url=%s, wsID=%s, msg=%s", resp.StatusCode, url, workspaceID, errBody.Message)
	}
	return nil
}

func (c *CrushServerClient) SendMessage(ctx context.Context, workspaceID, prompt, sessionID string) error {
	body := map[string]any{
		"prompt": prompt,
	}
	if sessionID != "" {
		body["session_id"] = sessionID
	}
	bodyData, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", c.host+"/v1/workspaces/"+workspaceID+"/agent", bytes.NewReader(bodyData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send message: status %d", resp.StatusCode)
	}
	return nil
}
