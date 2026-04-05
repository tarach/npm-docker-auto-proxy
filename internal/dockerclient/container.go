package dockerclient

import (
	"context"
	"net/url"
	"strings"
)

type ContainerInfo struct {
	ID     string
	Name   string
	Labels map[string]string
	State  string
}

type containerListItem struct {
	ID     string            `json:"Id"`
	Names  []string          `json:"Names"`
	Labels map[string]string `json:"Labels"`
	State  string            `json:"State"`
}

type containerInspectResponse struct {
	ID     string `json:"Id"`
	Name   string `json:"Name"`
	Config struct {
		Labels map[string]string `json:"Labels"`
	} `json:"Config"`
	State struct {
		Status string `json:"Status"`
	} `json:"State"`
}

func (c *Client) ListRunningContainers(ctx context.Context) ([]ContainerInfo, error) {
	var items []containerListItem

	err := c.doJSON(ctx, "GET", "/containers/json?all=false", &items)
	if err != nil {
		return nil, err
	}

	result := make([]ContainerInfo, 0, len(items))

	for _, item := range items {
		result = append(result, ContainerInfo{
			ID:     item.ID,
			Name:   cleanContainerName(item.Names),
			Labels: item.Labels,
			State:  item.State,
		})
	}

	return result, nil
}

func (c *Client) InspectContainer(ctx context.Context, containerID string) (ContainerInfo, error) {
	var item containerInspectResponse
	escapedID := url.PathEscape(containerID)

	err := c.doJSON(ctx, "GET", "/containers/"+escapedID+"/json", &item)
	if err != nil {
		return ContainerInfo{}, err
	}

	return ContainerInfo{
		ID:     item.ID,
		Name:   strings.TrimPrefix(item.Name, "/"),
		Labels: item.Config.Labels,
		State:  item.State.Status,
	}, nil
}

func cleanContainerName(names []string) string {
	if len(names) == 0 {
		return ""
	}

	return strings.TrimPrefix(names[0], "/")
}
