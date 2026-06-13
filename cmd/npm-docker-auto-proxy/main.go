package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tarach/npm-docker-auto-proxy/internal/config"
	"github.com/tarach/npm-docker-auto-proxy/internal/dockerclient"
	"github.com/tarach/npm-docker-auto-proxy/internal/logging"
	"github.com/tarach/npm-docker-auto-proxy/internal/npm"
	"github.com/tarach/npm-docker-auto-proxy/internal/syncer"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stdout, nil)).Error("config load failed", "event", "config_load_failed", "error", err.Error())
		os.Exit(1)
	}

	logger := logging.New(cfg.LogLevel)
	logger.Info("application started", "event", "app_started", "app", "npm-docker-auto-proxy")

	dockerClient := dockerclient.New(cfg.DockerSocketPath, logger)

	err = dockerClient.Ping(ctx)
	if err != nil {
		logger.Error("docker ping failed", "event", "docker_ping_failed", "socket", cfg.DockerSocketPath, "error", err.Error())
		os.Exit(1)
	}

	npmClient := npm.NewClient(cfg.NPMBaseURL, cfg.NPMEmail, cfg.NPMPassword, logger)

	err = npmClient.Login(ctx)
	if err != nil {
		logger.Error("npm login failed", "event", "npm_login_failed", "base_url", cfg.NPMBaseURL, "error", err.Error())
		os.Exit(1)
	}

	service := syncer.New(dockerClient, npmClient, logger, cfg.LabelsPrefix)

	err = service.SyncRunningContainers(ctx)
	if err != nil {
		logger.Error("initial container sync failed", "event", "initial_sync_failed", "error", err.Error())
		os.Exit(1)
	}

	err = service.WatchDockerEvents(ctx)
	if err != nil {
		logger.Error("docker events watcher stopped with error", "event", "docker_events_failed", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("application stopped", "event", "app_stopped")
}
