package npm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	email      string
	password   string
	token      string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewClient(baseURL string, email string, password string, logger *slog.Logger) *Client {
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		email:    email,
		password: password,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) doJSON(ctx context.Context, method string, path string, body any, target any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		c.logger.Error("npm api error", "event", "npm_api_error", "method", method, "path", path, "status", resp.StatusCode, "body", string(raw))
		return fmt.Errorf("npm api %s %s failed with status %d", method, path, resp.StatusCode)
	}

	if target == nil {
		return nil
	}

	if len(raw) == 0 {
		return nil
	}

	return json.Unmarshal(raw, target)
}

func (c *Client) doNoBody(ctx context.Context, method string, path string, target any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return err
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		c.logger.Error("npm api error", "event", "npm_api_error", "method", method, "path", path, "status", resp.StatusCode, "body", string(raw))
		return fmt.Errorf("npm api %s %s failed with status %d", method, path, resp.StatusCode)
	}

	if target == nil {
		return nil
	}

	if len(raw) == 0 {
		return nil
	}

	return json.Unmarshal(raw, target)
}
