# npm-docker-auto-proxy

Small Go daemon that watches Docker container events and creates, updates, enables, disables, or deletes Nginx Proxy Manager proxy hosts through the NPM API.

## Current MVP

- Go daemon
- Docker events
- Initial scan of already running containers
- NPM API login
- Create proxy hosts
- Update proxy hosts
- Enable proxy hosts on container start
- Disable or delete proxy hosts on container stop
- JSON logs with `log/slog`
- Code style avoids `switch`, `else`, and `else if`

## Environment variables

```env
NPM_BASE_URL=http://nginx-proxy-manager:81/api
NPM_EMAIL=admin@example.com
NPM_PASSWORD=change-me
LOG_LEVEL=info
DOCKER_SOCKET_PATH=/var/run/docker.sock
```

`LOG_LEVEL` accepts:

```text
debug
info
warn
error
```

## Labels

```yaml
labels:
  npm.proxy.enabled: "true"
  npm.proxy.domain: "jellyfin.tarach.net"
  npm.proxy.forward_host: "jellyfin"
  npm.proxy.forward_port: "8096"
  npm.proxy.scheme: "http"
  npm.proxy.websocket: "true"
  npm.proxy.ssl: "false"
  npm.proxy.force_ssl: "false"
  npm.proxy.http2: "true"
  npm.proxy.block_exploits: "true"
  npm.proxy.on_stop: "disable"
```

## Stop behavior

```yaml
npm.proxy.on_stop: "delete"
npm.proxy.on_stop: "del"
```

Deletes the proxy host.

```yaml
npm.proxy.on_stop: "disable"
npm.proxy.on_stop: "dis"
npm.proxy.on_stop: "off"
```

Disables the proxy host.

Missing `npm.proxy.on_stop` leaves NPM unchanged on container stop.

## Safety rule

Stop behavior works only when `npm.proxy.enabled=true` and a valid domain is configured.

## Build

```bash
docker compose build --no-cache --progress=plain
```

## Run

```bash
docker compose up -d
```

## Logs

```bash
docker logs npm-docker-auto-proxy
```

Pretty JSON:

```bash
docker logs npm-docker-auto-proxy | jq
```

## Local Go run

```bash
export NPM_BASE_URL="http://192.168.1.110:30020/api"
export NPM_EMAIL="admin@example.com"
export NPM_PASSWORD="change-me"
export LOG_LEVEL="debug"
export DOCKER_SOCKET_PATH="/var/run/docker.sock"

go run ./cmd/npm-docker-auto-proxy
```
