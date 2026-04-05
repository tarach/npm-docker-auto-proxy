package dockerclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	socketPath string
	httpClient *http.Client
	logger     *slog.Logger
}

func New(socketPath string, logger *slog.Logger) *Client {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network string, address string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socketPath)
		},
	}

	return &Client{
		socketPath: socketPath,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   0,
		},
		logger: logger,
	}
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := c.newRequest(ctx, "GET", "/_ping")
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("docker ping failed with status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) doJSON(ctx context.Context, method string, path string, target any) error {
	req, err := c.newRequest(ctx, method, path)
	if err != nil {
		return err
	}

	client := c.clientForPath(path)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("docker api %s %s failed with status %d: %s", method, path, resp.StatusCode, string(raw))
	}

	if target == nil {
		return nil
	}

	return json.Unmarshal(raw, target)
}

func (c *Client) newRequest(ctx context.Context, method string, path string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, "http://docker"+path, nil)
}

func (c *Client) clientForPath(path string) *http.Client {
	parsed, err := url.Parse(path)
	if err != nil {
		return c.shortRequestClient()
	}

	if parsed.Path == "/events" {
		return c.httpClient
	}

	return c.shortRequestClient()
}

func (c *Client) shortRequestClient() *http.Client {
	return &http.Client{
		Transport: c.httpClient.Transport,
		Timeout:   15 * time.Second,
	}
}
