package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultSidecarTimeout = 10 * time.Second
)

type SidecarClient struct {
	baseURL    string
	httpClient *http.Client
}

type SidecarHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type SidecarCommandRequest struct {
	Prompt        string   `json:"prompt"`
	CurrentFile   string   `json:"current_file,omitempty"`
	AllowedRoots  []string `json:"allowed_roots,omitempty"`
	DefaultOutput string   `json:"default_output,omitempty"`
	Provider      string   `json:"provider,omitempty"`
	Model         string   `json:"model,omitempty"`
}

type SidecarCommandResponse struct {
	ResultText  string `json:"result_text"`
	CurrentFile string `json:"current_file,omitempty"`
	Error       string `json:"error,omitempty"`
}

func NewSidecarClient(baseURL string, timeout time.Duration) *SidecarClient {
	baseURL = strings.TrimSpace(baseURL)
	baseURL = strings.TrimRight(baseURL, "/")
	if timeout <= 0 {
		timeout = DefaultSidecarTimeout
	}
	return &SidecarClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *SidecarClient) IsConfigured() bool {
	return c != nil && strings.TrimSpace(c.baseURL) != ""
}

func (c *SidecarClient) Health(ctx context.Context) (SidecarHealth, error) {
	if !c.IsConfigured() {
		return SidecarHealth{}, fmt.Errorf("sidecar base_url gerekli")
	}

	var health SidecarHealth
	if err := c.getJSON(ctx, "/health", &health); err != nil {
		return SidecarHealth{}, err
	}
	if strings.TrimSpace(health.Status) == "" {
		health.Status = "ok"
	}
	return health, nil
}

func (c *SidecarClient) RunCommand(ctx context.Context, req SidecarCommandRequest) (SidecarCommandResponse, error) {
	if !c.IsConfigured() {
		return SidecarCommandResponse{}, fmt.Errorf("sidecar base_url gerekli")
	}
	if strings.TrimSpace(req.Prompt) == "" {
		return SidecarCommandResponse{}, fmt.Errorf("prompt zorunlu")
	}

	var res SidecarCommandResponse
	if err := c.postJSON(ctx, "/v1/ai/command", req, &res); err != nil {
		return SidecarCommandResponse{}, err
	}
	if strings.TrimSpace(res.Error) != "" {
		return SidecarCommandResponse{}, errors.New(res.Error)
	}
	return res, nil
}

func (c *SidecarClient) getJSON(ctx context.Context, path string, out any) error {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeHTTPResponse(resp, out)
}

func (c *SidecarClient) postJSON(ctx context.Context, path string, payload any, out any) error {
	url := c.baseURL + path
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeHTTPResponse(resp, out)
}

func decodeHTTPResponse(resp *http.Response, out any) error {
	const maxErrorBody = 512
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return fmt.Errorf("sidecar http %d: %s", resp.StatusCode, msg)
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return err
	}
	return nil
}
