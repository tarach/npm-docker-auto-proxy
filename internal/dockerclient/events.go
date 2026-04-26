package dockerclient

import (
	"context"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
)

type ContainerEvent struct {
	ID         string
	Name       string
	Action     string
	Attributes map[string]string
}

func (c *Client) WatchContainerEvents(ctx context.Context) (<-chan ContainerEvent, <-chan error) {
	filter := filters.NewArgs()
	filter.Add("type", "container")

	messages, errs := c.api.Events(ctx, events.ListOptions{Filters: filter})
	out := make(chan ContainerEvent)

	go func() {
		defer close(out)

		for message := range messages {
			out <- ContainerEvent{
				ID:         message.Actor.ID,
				Name:       message.Actor.Attributes["name"],
				Action:     message.Action,
				Attributes: message.Actor.Attributes,
			}
		}
	}()

	return out, errs
}
