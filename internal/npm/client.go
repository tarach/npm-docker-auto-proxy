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

type npmErrorResponse struct {
	Error struct {
		Code        int    `json:"code"`
		Message     string `json:"message"`
		MessageI18N struct {
			Name      string `json:"name"`
			Message   string `json:"message"`
			ExpiredAt string `json:"expiredAt"`
		} `json:"message_i18n"`
	} `json:"error"`
}

type requestOptions struct {
	includeAuth     bool
	retryExpiredJWT bool
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
	return c.doJSONWithOptions(ctx, method, path, body, target, requestOptions{
		includeAuth:     true,
		retryExpiredJWT: true,
	})
}

func (c *Client) doJSONNoAuth(ctx context.Context, method string, path string, body any, target any) error {
	return c.doJSONWithOptions(ctx, method, path, body, target, requestOptions{
		includeAuth:     false,
		retryExpiredJWT: false,
	})
}

func (c *Client) doJSONWithOptions(ctx context.Context, method string, path string, body any, target any, options requestOptions) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(payload)), nil
	}

	return c.doRequest(req, method, path, target, options)
}

func (c *Client) doNoBody(ctx context.Context, method string, path string, target any) error {
	return c.doNoBodyWithOptions(ctx, method, path, target, requestOptions{
		includeAuth:     true,
		retryExpiredJWT: true,
	})
}

func (c *Client) doNoBodyWithOptions(ctx context.Context, method string, path string, target any, options requestOptions) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req, method, path, target, options)
}

func (c *Client) doRequest(req *http.Request, method string, path string, target any, options requestOptions) error {
	c.applyAuthorization(req, options)

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
		return c.handleAPIError(req, method, path, target, options, resp.StatusCode, raw)
	}

	return decodeResponse(raw, target)
}

func (c *Client) handleAPIError(req *http.Request, method string, path string, target any, options requestOptions, statusCode int, raw []byte) error {
	if options.retryExpiredJWT && isTokenExpired(raw) {
		c.logger.Warn(
			"npm token expired, refreshing token",
			"event", "npm_token_expired",
			"method", method,
			"path", path,
		)

		err := c.Login(req.Context())
		if err != nil {
			return err
		}

		return c.retryRequest(req, method, path, target, requestOptions{
			includeAuth:     options.includeAuth,
			retryExpiredJWT: false,
		})
	}

	c.logger.Error(
		"npm api error",
		"event", "npm_api_error",
		"method", method,
		"path", path,
		"status", statusCode,
		"body", string(raw),
	)

	return fmt.Errorf("npm api %s %s failed with status %d", method, path, statusCode)
}

func (c *Client) retryRequest(req *http.Request, method string, path string, target any, options requestOptions) error {
	cloned, err := cloneRequest(req)
	if err != nil {
		return err
	}

	return c.doRequest(cloned, method, path, target, options)
}

func cloneRequest(req *http.Request) (*http.Request, error) {
	cloned := req.Clone(req.Context())
	cloned.Header = req.Header.Clone()

	if req.GetBody == nil {
		return cloned, nil
	}

	body, err := req.GetBody()
	if err != nil {
		return nil, err
	}

	cloned.Body = body

	return cloned, nil
}

func (c *Client) applyAuthorization(req *http.Request, options requestOptions) {
	if !options.includeAuth {
		return
	}

	if c.token == "" {
		return
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
}

func decodeResponse(raw []byte, target any) error {
	if target == nil {
		return nil
	}

	if len(raw) == 0 {
		return nil
	}

	return json.Unmarshal(raw, target)
}

func isTokenExpired(raw []byte) bool {
	var response npmErrorResponse

	err := json.Unmarshal(raw, &response)
	if err != nil {
		return false
	}

	matches := []bool{
		response.Error.Message == "Token has expired",
		response.Error.MessageI18N.Name == "TokenExpiredError",
		response.Error.MessageI18N.Message == "jwt expired",
	}

	for _, match := range matches {
		if match {
			return true
		}
	}

	return false
}
