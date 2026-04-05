package dockerclient

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

type ContainerEvent struct {
	ID         string
	Name       string
	Action     string
	Attributes map[string]string
}

type ContainerEventResult struct {
	Event ContainerEvent
	Err   error
}

type eventMessage struct {
	Type   string `json:"Type"`
	Action string `json:"Action"`
	Actor  struct {
		ID         string            `json:"ID"`
		Attributes map[string]string `json:"Attributes"`
	} `json:"Actor"`
}

func (c *Client) WatchContainerEvents(ctx context.Context) <-chan ContainerEventResult {
	out := make(chan ContainerEventResult)

	go func() {
		defer close(out)

		req, err := c.newRequest(ctx, "GET", "/events?filters="+containerEventFilter())
		if err != nil {
			out <- ContainerEventResult{Err: err}
			return
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			out <- ContainerEventResult{Err: err}
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			out <- ContainerEventResult{
				Err: fmt.Errorf("docker events failed with status %d", resp.StatusCode),
			}
			return
		}

		c.readEvents(ctx, resp.Body, out)
	}()

	return out
}

func (c *Client) readEvents(ctx context.Context, body io.Reader, out chan<- ContainerEventResult) {
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		message, err := decodeEventMessage(scanner.Bytes())
		if err != nil {
			c.logger.Warn(
				"docker event decode failed",
				"event", "docker_event_decode_failed",
				"error", err.Error(),
			)
			continue
		}

		if message.Type != "container" {
			continue
		}

		out <- ContainerEventResult{
			Event: ContainerEvent{
				ID:         message.Actor.ID,
				Name:       message.Actor.Attributes["name"],
				Action:     message.Action,
				Attributes: message.Actor.Attributes,
			},
		}
	}

	err := scanner.Err()
	if err != nil {
		out <- ContainerEventResult{Err: err}
		return
	}

	if ctx.Err() != nil {
		return
	}

	out <- ContainerEventResult{
		Err: fmt.Errorf("docker events stream ended"),
	}
}

func decodeEventMessage(raw []byte) (eventMessage, error) {
	var message eventMessage

	err := json.Unmarshal(raw, &message)
	if err != nil {
		return eventMessage{}, err
	}

	return message, nil
}

func containerEventFilter() string {
	return url.QueryEscape(`{
		"type": ["container"],
		"event": ["create", "start", "restart", "die", "stop", "destroy"]
	}`)
}
