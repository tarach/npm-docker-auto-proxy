package syncer

import (
	"context"
	"log/slog"

	"github.com/tarach/npm-docker-auto-proxy/internal/dockerclient"
	"github.com/tarach/npm-docker-auto-proxy/internal/npm"
	"github.com/tarach/npm-docker-auto-proxy/internal/proxy"
)

type Docker interface {
	ListRunningContainers(ctx context.Context) ([]dockerclient.ContainerInfo, error)
	InspectContainer(ctx context.Context, containerID string) (dockerclient.ContainerInfo, error)
	WatchContainerEvents(ctx context.Context) (<-chan dockerclient.ContainerEvent, <-chan error)
}

type NPM interface {
	FindProxyHostByDomain(ctx context.Context, domain string) (npm.ProxyHost, bool, error)
	CreateProxyHost(ctx context.Context, desired proxy.DesiredHost) (npm.ProxyHost, error)
	UpdateProxyHost(ctx context.Context, id int, desired proxy.DesiredHost) (npm.ProxyHost, error)
	EnableProxyHost(ctx context.Context, id int) error
	DisableProxyHost(ctx context.Context, id int) error
	DeleteProxyHost(ctx context.Context, id int) error
}

type Syncer struct {
	docker Docker
	npm    NPM
	logger *slog.Logger
}

func New(docker Docker, npm NPM, logger *slog.Logger) *Syncer {
	return &Syncer{
		docker: docker,
		npm:    npm,
		logger: logger,
	}
}

func (s *Syncer) SyncRunningContainers(ctx context.Context) error {
	containers, err := s.docker.ListRunningContainers(ctx)
	if err != nil {
		return err
	}

	s.logger.Info("initial container scan started", "event", "initial_scan_started", "containers", len(containers))

	for _, container := range containers {
		err := s.HandleContainerStart(ctx, container)
		if err != nil {
			s.logger.Error("container sync failed", "event", "container_sync_failed", "container", container.Name, "container_id", container.ID, "error", err.Error())
		}
	}

	s.logger.Info("initial container scan finished", "event", "initial_scan_finished")

	return nil
}

func (s *Syncer) WatchDockerEvents(ctx context.Context) error {
	events, errs := s.docker.WatchContainerEvents(ctx)

	startActions := map[string]func(context.Context, dockerclient.ContainerEvent) error{
		"create":  s.handleStartEvent,
		"start":   s.handleStartEvent,
		"restart": s.handleStartEvent,
	}

	stopActions := map[string]func(context.Context, dockerclient.ContainerEvent) error{
		"die":     s.handleStopEvent,
		"stop":    s.handleStopEvent,
		"destroy": s.handleStopEvent,
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errs:
			if err != nil {
				return err
			}
		case event, ok := <-events:
			if !ok {
				return nil
			}

			s.logger.Info("docker event received", "event", "docker_event_received", "docker_action", event.Action, "container", event.Name, "container_id", event.ID)

			handler, found := startActions[event.Action]
			if found {
				err := handler(ctx, event)
				if err != nil {
					s.logger.Error("start event handling failed", "event", "start_event_failed", "docker_action", event.Action, "container", event.Name, "container_id", event.ID, "error", err.Error())
				}

				continue
			}

			handler, found = stopActions[event.Action]
			if found {
				err := handler(ctx, event)
				if err != nil {
					s.logger.Error("stop event handling failed", "event", "stop_event_failed", "docker_action", event.Action, "container", event.Name, "container_id", event.ID, "error", err.Error())
				}

				continue
			}

			s.logger.Debug("docker event ignored", "event", "docker_event_ignored", "docker_action", event.Action, "container", event.Name, "container_id", event.ID)
		}
	}
}

func (s *Syncer) handleStartEvent(ctx context.Context, event dockerclient.ContainerEvent) error {
	container, err := s.docker.InspectContainer(ctx, event.ID)
	if err != nil {
		return err
	}

	return s.HandleContainerStart(ctx, container)
}

func (s *Syncer) handleStopEvent(ctx context.Context, event dockerclient.ContainerEvent) error {
	labels := labelsFromEvent(event)
	container := dockerclient.ContainerInfo{
		ID:     event.ID,
		Name:   event.Name,
		Labels: labels,
	}

	return s.HandleContainerStop(ctx, container)
}

func (s *Syncer) HandleContainerStart(ctx context.Context, container dockerclient.ContainerInfo) error {
	desired, err := proxy.FromLabels(container.Labels)
	if err != nil {
		s.logger.Warn("container has invalid proxy labels", "event", "container_invalid_labels", "container", container.Name, "container_id", container.ID, "error", err.Error())
		return nil
	}

	if !desired.Enabled {
		s.logger.Debug("container skipped", "event", "container_skipped", "reason", "proxy_not_enabled", "container", container.Name, "container_id", container.ID)
		return nil
	}

	existing, found, err := s.npm.FindProxyHostByDomain(ctx, desired.Domain)
	if err != nil {
		return err
	}

	if !found {
		created, err := s.npm.CreateProxyHost(ctx, desired)
		if err != nil {
			return err
		}

		s.logger.Info("proxy host created", "event", "proxy_host_created", "container", container.Name, "container_id", container.ID, "domain", desired.Domain, "proxy_host_id", created.ID, "forward_host", desired.ForwardHost, "forward_port", desired.ForwardPort)
		return nil
	}

	if !existing.Enabled {
		err := s.npm.EnableProxyHost(ctx, existing.ID)
		if err != nil {
			return err
		}

		s.logger.Info("proxy host enabled", "event", "proxy_host_enabled", "container", container.Name, "container_id", container.ID, "domain", desired.Domain, "proxy_host_id", existing.ID)
	}

	changes := npm.Differences(existing, desired)
	if len(changes) == 0 {
		s.logger.Info("proxy host unchanged", "event", "proxy_host_unchanged", "container", container.Name, "container_id", container.ID, "domain", desired.Domain, "proxy_host_id", existing.ID)
		return nil
	}

	updated, err := s.npm.UpdateProxyHost(ctx, existing.ID, desired)
	if err != nil {
		return err
	}

	s.logger.Info("proxy host updated", "event", "proxy_host_updated", "container", container.Name, "container_id", container.ID, "domain", desired.Domain, "proxy_host_id", updated.ID, "changed", changes)

	return nil
}

func (s *Syncer) HandleContainerStop(ctx context.Context, container dockerclient.ContainerInfo) error {
	desired, err := proxy.FromLabels(container.Labels)
	if err != nil {
		s.logger.Warn("container has invalid proxy labels on stop", "event", "container_invalid_labels_on_stop", "container", container.Name, "container_id", container.ID, "error", err.Error())
		return nil
	}

	if !desired.Enabled {
		s.logger.Debug("stop ignored", "event", "container_stop_ignored", "reason", "proxy_not_enabled", "container", container.Name, "container_id", container.ID)
		return nil
	}

	if desired.OnStop == proxy.StopActionNone {
		s.logger.Info("stop ignored", "event", "container_stop_ignored", "reason", "on_stop_not_configured", "container", container.Name, "container_id", container.ID, "domain", desired.Domain)
		return nil
	}

	existing, found, err := s.npm.FindProxyHostByDomain(ctx, desired.Domain)
	if err != nil {
		return err
	}

	if !found {
		s.logger.Info("stop action skipped", "event", "stop_action_skipped", "reason", "proxy_host_not_found", "container", container.Name, "container_id", container.ID, "domain", desired.Domain, "action", desired.OnStop)
		return nil
	}

	handlers := map[proxy.StopAction]func(context.Context, int) error{
		proxy.StopActionDelete:  s.npm.DeleteProxyHost,
		proxy.StopActionDisable: s.npm.DisableProxyHost,
	}

	handler, found := handlers[desired.OnStop]
	if !found {
		s.logger.Warn("unknown on stop action", "event", "unknown_on_stop_action", "container", container.Name, "container_id", container.ID, "domain", desired.Domain, "action", desired.OnStop)
		return nil
	}

	err = handler(ctx, existing.ID)
	if err != nil {
		return err
	}

	s.logStopAction(container, desired, existing.ID)

	return nil
}

func (s *Syncer) logStopAction(container dockerclient.ContainerInfo, desired proxy.DesiredHost, proxyHostID int) {
	events := map[proxy.StopAction]string{
		proxy.StopActionDelete:  "proxy_host_deleted",
		proxy.StopActionDisable: "proxy_host_disabled",
	}

	messages := map[proxy.StopAction]string{
		proxy.StopActionDelete:  "proxy host deleted",
		proxy.StopActionDisable: "proxy host disabled",
	}

	s.logger.Info(messages[desired.OnStop], "event", events[desired.OnStop], "container", container.Name, "container_id", container.ID, "domain", desired.Domain, "proxy_host_id", proxyHostID, "reason", "container_stopped")
}

func labelsFromEvent(event dockerclient.ContainerEvent) map[string]string {
	labels := map[string]string{}

	for key, value := range event.Attributes {
		labels[key] = value
	}

	return labels
}
