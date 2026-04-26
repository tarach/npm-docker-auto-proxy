package dockerclient

import (
	"context"

	"github.com/docker/docker/client"
)

type Client struct {
	api *client.Client
}

func New() (*Client, error) {
	api, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &Client{api: api}, nil
}

func (c *Client) Close() error {
	return c.api.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.api.Ping(ctx)
	return err
}
