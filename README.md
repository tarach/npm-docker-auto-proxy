# npm-docker-auto-proxy

Docker image: https://hub.docker.com/r/tarach/npm-docker-auto-proxy  
Source code: https://github.com/tarach/npm-docker-auto-proxy  
NPM Repo: https://github.com/NginxProxyManager/nginx-proxy-manager

## TL;DR

`npm-docker-auto-proxy` adds Traefik-like Docker labels to Nginx Proxy Manager.

You keep using NPM as your reverse proxy and UI, but proxy hosts can be created, updated, disabled, or deleted automatically from Docker container labels.

Good for homelab / TrueNAS SCALE / Docker Compose setups where you do not want to manually click through NPM every time you add a service.

### Why?

Nginx Proxy Manager is great when you want a UI for reverse proxy hosts.

But if you run many Docker services, manually creating a proxy host for every container gets repetitive.

This project keeps NPM in place and adds a small automation layer on top of it:
Docker labels in, NPM proxy hosts out.

### Example
```yaml
services:
  jellyfin:
    image: jellyfin/jellyfin
    labels:
      npm.proxy.enabled: "true"
      npm.proxy.domain: "jellyfin.example.com"
      npm.proxy.forward_host: "jellyfin"
      npm.proxy.forward_port: "8096"
      npm.proxy.scheme: "http"
      npm.proxy.ssl: "true"
      npm.proxy.certificate: "*.example.com"
      npm.proxy.force_ssl: "true"
```
When this container starts, `npm-docker-auto-proxy` creates or updates the matching NPM proxy host automatically.

### What happens on Docker events?

| Docker event | NPM action |
|---|---|
| Container starts | Creates, updates, and enables the matching proxy host |
| Container stops | Disables, deletes, or keeps the proxy host, depending on `npm.proxy.on_stop` |
| Container is recreated with changed labels | Updates the matching proxy host |
| `npm-docker-auto-proxy` starts | Runs an initial scan and catches already-running containers |

## Table of contents

- [TL;DR](https://github.com/tarach/npm-docker-auto-proxy#tldr)
  - [Why?](https://github.com/tarach/npm-docker-auto-proxy#why)
  - [Example](https://github.com/tarach/npm-docker-auto-proxy#example)
  - [What happens on Docker events?](https://github.com/tarach/npm-docker-auto-proxy#what-happens-on-docker-events)
- [Features](https://github.com/tarach/npm-docker-auto-proxy#features)
- [How it works](https://github.com/tarach/npm-docker-auto-proxy#how-it-works)
- [Docker socket access](https://github.com/tarach/npm-docker-auto-proxy#docker-socket-access)
- [Environment variables](https://github.com/tarach/npm-docker-auto-proxy#environment-variables)
  - [`NPM_BASE_URL`](https://github.com/tarach/npm-docker-auto-proxy#npm_base_url)
  - [`NPM_EMAIL`](https://github.com/tarach/npm-docker-auto-proxy#npm_email)
  - [`NPM_PASSWORD`](https://github.com/tarach/npm-docker-auto-proxy#npm_password)
  - [`LOG_LEVEL`](https://github.com/tarach/npm-docker-auto-proxy#log_level)
  - [`DOCKER_SOCKET_PATH`](https://github.com/tarach/npm-docker-auto-proxy#docker_socket_path)
  - [`LABELS_PREFIX`](https://github.com/tarach/npm-docker-auto-proxy#labels_prefix)
- [Docker Compose](https://github.com/tarach/npm-docker-auto-proxy#docker-compose)
- [Build](https://github.com/tarach/npm-docker-auto-proxy#build)
- [Run](https://github.com/tarach/npm-docker-auto-proxy#run)
- [Logs](https://github.com/tarach/npm-docker-auto-proxy#logs)
- [Container labels](https://github.com/tarach/npm-docker-auto-proxy#container-labels)
- [TrueNAS Scale](https://github.com/tarach/npm-docker-auto-proxy#truenas-scale)
  - [Configure labels in the TrueNAS UI](https://github.com/tarach/npm-docker-auto-proxy#configure-labels-in-the-truenas-ui)
- [Supported labels](https://github.com/tarach/npm-docker-auto-proxy#supported-labels)
  - [`npm.proxy.enabled`](https://github.com/tarach/npm-docker-auto-proxy#npmproxyenabled)
  - [`npm.proxy.domain`](https://github.com/tarach/npm-docker-auto-proxy#npmproxydomain)
  - [`npm.proxy.forward_host`](https://github.com/tarach/npm-docker-auto-proxy#npmproxyforward_host)
  - [`npm.proxy.forward_port`](https://github.com/tarach/npm-docker-auto-proxy#npmproxyforward_port)
  - [`npm.proxy.scheme`](https://github.com/tarach/npm-docker-auto-proxy#npmproxyscheme)
  - [`npm.proxy.websocket`](https://github.com/tarach/npm-docker-auto-proxy#npmproxywebsocket)
  - [`npm.proxy.block_exploits`](https://github.com/tarach/npm-docker-auto-proxy#npmproxyblock_exploits)
  - [`npm.proxy.http2`](https://github.com/tarach/npm-docker-auto-proxy#npmproxyhttp2)
- [SSL certificates](https://github.com/tarach/npm-docker-auto-proxy#ssl-certificates)
  - [Option 1: certificate name or domain](https://github.com/tarach/npm-docker-auto-proxy#option-1-certificate-name-or-domain)
  - [Option 2: certificate ID](https://github.com/tarach/npm-docker-auto-proxy#option-2-certificate-id)
  - [Certificate lookup permissions](https://github.com/tarach/npm-docker-auto-proxy#certificate-lookup-permissions)
  - [SSL validation rules](https://github.com/tarach/npm-docker-auto-proxy#ssl-validation-rules)
- [Stop behavior](https://github.com/tarach/npm-docker-auto-proxy#stop-behavior)
  - [Delete on stop](https://github.com/tarach/npm-docker-auto-proxy#delete-on-stop)
  - [Disable on stop](https://github.com/tarach/npm-docker-auto-proxy#disable-on-stop)
  - [No stop action](https://github.com/tarach/npm-docker-auto-proxy#no-stop-action)
- [Safety rules](https://github.com/tarach/npm-docker-auto-proxy#safety-rules)
- [Example: Jellyfin](https://github.com/tarach/npm-docker-auto-proxy#example-jellyfin)
- [Testing NPM API access](https://github.com/tarach/npm-docker-auto-proxy#testing-npm-api-access)
- [Testing backend reachability from NPM](https://github.com/tarach/npm-docker-auto-proxy#testing-backend-reachability-from-npm)
- [Development](https://github.com/tarach/npm-docker-auto-proxy#development)
- [Notes](https://github.com/tarach/npm-docker-auto-proxy#notes)

## Features

- Watches Docker container events.
- Performs an initial scan of already running containers.
- Creates Nginx Proxy Manager proxy hosts.
- Updates existing proxy hosts when labels change.
- Enables proxy hosts when containers start.
- Disables or deletes proxy hosts when containers stop.
- Supports SSL certificates by certificate name/domain or certificate ID.
- Supports a configurable Docker label prefix through `LABELS_PREFIX`.
- Uses structured JSON logs through Go `log/slog`.
- Uses Docker Engine HTTP API through `/var/run/docker.sock`.

## How it works

The application listens to Docker container events:

```text
Docker container events
        ↓
npm-docker-auto-proxy
        ↓
Docker labels
        ↓
Nginx Proxy Manager API
        ↓
Proxy Host create/update/enable/disable/delete
```

On startup, it scans all currently running containers.

After that, it listens for selected Docker events:

```text
create
start
restart
die
stop
destroy
```

Other Docker events are ignored.

## Docker socket access

The application needs access to the Docker socket:

```yaml
volumes:
  - /var/run/docker.sock:/var/run/docker.sock:ro
```

Access to `/var/run/docker.sock` is powerful. Even when mounted as read-only, the Docker API exposed by the socket can still allow privileged Docker operations depending on permissions.

For a first test, the container can run as root:

```yaml
user: "0:0"
```

A safer setup is to run the container as a non-root user with access to the group that owns `/var/run/docker.sock`.

Check Docker socket ownership:

```bash
stat -c 'user=%U group=%G uid=%u gid=%g %A %n' /var/run/docker.sock
```

Example:

```text
user=root group=docker uid=0 gid=999 srw-rw---- /var/run/docker.sock
```

This means the socket is owned by `root:docker`, and the Docker socket group ID is `999`.

To run the container as a non-root user, add that group ID to the container:

```yaml
group_add:
  - "999"
```

## Environment variables

```env
NPM_BASE_URL=http://nginx-proxy-manager:81/api
NPM_EMAIL=admin@example.com
NPM_PASSWORD=change-me
LOG_LEVEL=info
DOCKER_SOCKET_PATH=/var/run/docker.sock
LABELS_PREFIX=npm.proxy.
```

### `NPM_BASE_URL`

Base URL of the Nginx Proxy Manager API.

Examples:

```env
NPM_BASE_URL=http://nginx-proxy-manager:81/api
```

or:

```env
NPM_BASE_URL=http://192.168.1.2:30020/api
```

### `NPM_EMAIL`

Nginx Proxy Manager API user email.

```env
NPM_EMAIL=auto.proxy@example.com
```

### `NPM_PASSWORD`

Nginx Proxy Manager API user password.

```env
NPM_PASSWORD=change-me
```

### `LOG_LEVEL`

Supported values:

```text
debug
info
warn
error
```

Default:

```env
LOG_LEVEL=info
```

Use debug while testing:

```env
LOG_LEVEL=debug
```

### `DOCKER_SOCKET_PATH`

Path to the Docker socket inside the container.

Default:

```env
DOCKER_SOCKET_PATH=/var/run/docker.sock
```

### `LABELS_PREFIX`

Prefix used for Docker label keys.

Default:

```env
LABELS_PREFIX=npm.proxy.
```

With the default prefix, a container uses labels such as `npm.proxy.enabled` and `npm.proxy.domain`.

To use a custom prefix:

```env
LABELS_PREFIX=myapp.proxy.
```

Then use matching labels on containers:

```yaml
labels:
  myapp.proxy.enabled: "true"
  myapp.proxy.domain: "app.example.com"
  myapp.proxy.forward_host: "app"
  myapp.proxy.forward_port: "8080"
```

If the prefix does not end with `.`, a trailing dot is added automatically. For example, `myapp.proxy` becomes `myapp.proxy.`.

The `LABELS_PREFIX` value in `npm-docker-auto-proxy` must match the prefix used on container labels.

## Docker Compose

Example `docker-compose.yml`:

```yaml
services:
  npm-docker-auto-proxy:
    image: tarach/npm-docker-auto-proxy:latest
    build: .
    container_name: npm-docker-auto-proxy
    restart: unless-stopped
    user: "0:0"
    environment:
      NPM_BASE_URL: "http://192.168.1.2:30020/api"
      NPM_EMAIL: "auto.proxy@example.com"
      NPM_PASSWORD: "change-me"
      LOG_LEVEL: "info"
      DOCKER_SOCKET_PATH: "/var/run/docker.sock"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    networks:
      - proxy

networks:
  proxy:
    external: true
```

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

Pretty JSON logs:

```bash
docker logs npm-docker-auto-proxy | jq
```

Follow logs with timestamps:

```bash
docker logs -tf npm-docker-auto-proxy
```

## Container labels

Containers declare proxy settings through Docker labels. By default, label keys start with `npm.proxy.` (see [`LABELS_PREFIX`](https://github.com/tarach/npm-docker-auto-proxy#labels_prefix)).

A proxied container must have:

```yaml
labels:
  npm.proxy.enabled: "true"
  npm.proxy.domain: "jellyfin.domain.com"
  npm.proxy.forward_host: "jellyfin"
  npm.proxy.forward_port: "8096"
```

Full example:

```yaml
labels:
  npm.proxy.enabled: "true"
  npm.proxy.domain: "jellyfin.domain.com"
  npm.proxy.forward_host: "jellyfin"
  npm.proxy.forward_port: "8096"
  npm.proxy.scheme: "http"
  npm.proxy.websocket: "true"

  npm.proxy.ssl: "true"
  npm.proxy.certificate: "*.domain.com"
  npm.proxy.force_ssl: "true"
  npm.proxy.http2: "true"

  npm.proxy.block_exploits: "true"
  npm.proxy.on_stop: "disable"
```

## TrueNAS Scale

TrueNAS Scale Apps expose container labels through the web UI instead of a compose file. Adding many labels by hand is slow and error-prone.

This repository includes Chrome DevTools scripts that automate filling the **Labels** section of a TrueNAS app edit screen. They were tested on **TrueNAS 25.10.3.1 (Goldeye)**. Other Scale versions may work, but UI changes can break the selectors.

Scripts:

```text
examples/TrueNAS App Labels Configurator.js
examples/truenas.js
```

- `TrueNAS App Labels Configurator.js` defines `labelsSetupFunc`, which adds each label key/value pair and assigns it to a container.
- `truenas.js` is an example call with sample values. Copy and edit it for your app.

### Configure labels in the TrueNAS UI

1. Open the TrueNAS web UI and go to **Apps**.
2. Open the app you want to configure and click **Edit**.
3. Scroll to the **Labels** section and leave that section visible on screen.
4. Open the browser developer tools:
   - Chrome / Edge: `F12` or `Ctrl+Shift+I` (`Cmd+Option+I` on macOS)
   - Firefox: `F12` or `Ctrl+Shift+I`
5. Open the **Console** tab.
6. Paste the contents of `examples/TrueNAS App Labels Configurator.js` and press Enter.
7. Edit and paste the contents of `examples/truenas.js`, adjusting:
   - `labelContainerName` — the container name shown in the TrueNAS Labels UI (for example `subdomain-proxy`)
   - `labels` — the proxy labels for that container (default prefix: `npm.proxy.*`; must match [`LABELS_PREFIX`](https://github.com/tarach/npm-docker-auto-proxy#labels_prefix) if customized)
8. Press Enter and wait until the console prints `Done.`
9. Review the filled labels in the UI, then save and redeploy the app as usual.

Example `truenas.js` values:

```javascript
const labelContainerName = "subdomain-proxy";

const labels = {
    "npm.proxy.enabled": "true",
    "npm.proxy.domain": "subdomain-proxy.domain.com",
    "npm.proxy.forward_host": "192.168.1.2",
    "npm.proxy.forward_port": "81",
    "npm.proxy.scheme": "http",
    "npm.proxy.websocket": "true",
    "npm.proxy.ssl": "true",
    "npm.proxy.certificate": "*.domain.com",
    "npm.proxy.force_ssl": "true",
    "npm.proxy.http2": "true",
    "npm.proxy.block_exploits": "true",
    "npm.proxy.on_stop": "disable",
};

labelsSetupFunc(labelContainerName, labels);
```

Notes:

- Run the script only on the app **Edit** screen while the **Labels** section is present.
- `labelContainerName` must match the container name shown in the TrueNAS dropdown for that label row.
- The script adds labels; it does not remove existing ones. Clear unwanted labels manually before running it if needed.
- After the script finishes, confirm the values in the UI and save the app so Docker receives the updated labels.
- If TrueNAS updates its UI, the selectors in `TrueNAS App Labels Configurator.js` may need adjustment.

## Supported labels

### `npm.proxy.enabled`

Enables automatic proxy management for this container.

```yaml
npm.proxy.enabled: "true"
```

If this label is missing or not `true`, the container is ignored.

This safety rule also applies to stop actions. A container without:

```yaml
npm.proxy.enabled: "true"
```

will not trigger proxy host disable/delete actions.

### `npm.proxy.domain`

Domain name for the NPM proxy host.

```yaml
npm.proxy.domain: "jellyfin.domain.com"
```

This becomes:

```yaml
"domain_names": ["jellyfin.domain.com"]
```

### `npm.proxy.forward_host`

Backend hostname or IP used by NPM.

```yaml
npm.proxy.forward_host: "jellyfin"
```

If NPM and the target container are on the same Docker network, this can be the container name or network alias.

Example with network alias:

```yaml
services:
  jellyfin:
    networks:
      proxy:
        aliases:
          - jellyfin
```

Then use:

```yaml
npm.proxy.forward_host: "jellyfin"
```

### `npm.proxy.forward_port`

Backend port used by NPM.

```yaml
npm.proxy.forward_port: "8096"
```

Use the internal container port, not necessarily the host-published port.

For example, Jellyfin usually listens on:

```text
8096
```

inside the container.

### `npm.proxy.scheme`

Backend scheme.

```yaml
npm.proxy.scheme: "http"
```

Default:

```text
http
```

For Jellyfin, this is usually:

```yaml
npm.proxy.scheme: "http"
```

Even if the public site uses HTTPS, the backend connection from NPM to Jellyfin is usually HTTP.

### `npm.proxy.websocket`

Enables WebSocket support.

```yaml
npm.proxy.websocket: "true"
```

This becomes:

```yaml
"allow_websocket_upgrade": true
```

### `npm.proxy.block_exploits`

Enables NPM block common exploits option.

```yaml
npm.proxy.block_exploits: "true"
```

Default:

```text
true
```

### `npm.proxy.http2`

Enables HTTP/2 support on the NPM proxy host.

```yaml
npm.proxy.http2: "true"
```

Default:

```text
true
```

## SSL certificates

There are two supported ways to assign an SSL certificate to a proxy host.

### Option 1: certificate name or domain

Use this when the NPM API user has permission to list certificates.

```yaml
labels:
  npm.proxy.ssl: "true"
  npm.proxy.certificate: "*.domain.com"
  npm.proxy.force_ssl: "true"
```

The application calls:

```text
GET /api/nginx/certificates
```

Then it tries to match `npm.proxy.certificate` against:

```text
- certificate nice_name
- certificate domain_names
```

Example certificate returned by NPM:

```json
{
  "id": 3,
  "provider": "letsencrypt",
  "nice_name": "domain.com, *.domain.com",
  "domain_names": ["*.domain.com", "domain.com"],
  "expires_on": "2026-08-14 13:25:28"
}
```

For this label:

```yaml
npm.proxy.certificate: "*.domain.com"
```

the application resolves:

```yaml
"certificate_id": 3
```

and sends that to Nginx Proxy Manager.

You can also match by root domain:

```yaml
npm.proxy.certificate: "domain.com"
```

or by nice name:

```yaml
npm.proxy.certificate: "domain.com, *.domain.com"
```

### Option 2: certificate ID

Use this when the NPM API user cannot list certificates or when you want to avoid certificate lookup.

```yaml
labels:
  npm.proxy.ssl: "true"
  npm.proxy.certificate_id: "3"
  npm.proxy.force_ssl: "true"
```

`npm.proxy.certificate_id` is passed directly to Nginx Proxy Manager as:

```yaml
"certificate_id": 3
```

This option does not require certificate lookup permissions.

### Certificate lookup permissions

Nginx Proxy Manager users may only see certificates they own, depending on user permissions.

If:

```bash
curl -s "http://NPM_HOST:PORT/api/nginx/certificates" \
  -H "Authorization: Bearer ${TOKEN}"
```

returns:

```json
[]
```

then the API user probably cannot see existing certificates.

In that case, either use:

```yaml
npm.proxy.certificate_id: "3"
```

or adjust the user permissions in NPM so the user can list the required certificate.

### SSL validation rules

When SSL is enabled:

```yaml
npm.proxy.ssl: "true"
```

one of these labels is required:

```yaml
npm.proxy.certificate: "*.domain.com"
```

or:

```yaml
npm.proxy.certificate_id: "3"
```

`npm.proxy.force_ssl=true` requires:

```yaml
npm.proxy.ssl: "true"
```

Invalid SSL configuration prevents proxy host create/update and is logged as `container_invalid_labels`.

## Stop behavior

Stop behavior is controlled by:

```yaml
npm.proxy.on_stop: "disable"
```

Supported values:

```text
delete
del
disable
dis
off
```

### Delete on stop

```yaml
npm.proxy.on_stop: "delete"
```

or:

```yaml
npm.proxy.on_stop: "del"
```

Deletes the proxy host from Nginx Proxy Manager when the container stops.

### Disable on stop

```yaml
npm.proxy.on_stop: "disable"
```

or:

```yaml
npm.proxy.on_stop: "dis"
```

or:

```yaml
npm.proxy.on_stop: "off"
```

Disables the proxy host when the container stops.

### No stop action

If `npm.proxy.on_stop` is missing, the proxy host is left unchanged when the container stops.

## Safety rules

The application only acts on containers with:

```yaml
npm.proxy.enabled: "true"
```

Stop actions also require:

```yaml
npm.proxy.enabled: "true"
```

This means that a container with only:

```yaml
npm.proxy.on_stop: "delete"
```

will not delete anything.

The application does not log:

```text
- NPM password
- Bearer token
- Authorization header
```

## Example: Jellyfin

```yaml
services:
  jellyfin:
    image: jellyfin/jellyfin
    container_name: jellyfin
    restart: unless-stopped
    volumes:
      - /mnt/fast/apps/jellyfin/config:/config
      - /mnt/tank/movies:/media
    networks:
      proxy:
        aliases:
          - jellyfin
    labels:
      npm.proxy.enabled: "true"
      npm.proxy.domain: "jellyfin.domain.com"
      npm.proxy.forward_host: "jellyfin"
      npm.proxy.forward_port: "8096"
      npm.proxy.scheme: "http"
      npm.proxy.websocket: "true"

      npm.proxy.ssl: "true"
      npm.proxy.certificate: "*.domain.com"
      npm.proxy.force_ssl: "true"
      npm.proxy.http2: "true"

      npm.proxy.block_exploits: "true"
      npm.proxy.on_stop: "disable"

networks:
  proxy:
    external: true
```

Alternative SSL setup using certificate ID:

```yaml
      npm.proxy.ssl: "true"
      npm.proxy.certificate_id: "3"
      npm.proxy.force_ssl: "true"
```

## Testing NPM API access

Get a token:

```bash
TOKEN=$(
  curl -s -X POST "http://192.168.1.2:30020/api/tokens" \
    -H "Content-Type: application/json" \
    -d '{
      "identity": "auto.proxy@example.com",
      "secret": "change-me"
    }' | jq -r '.token'
)
```

List proxy hosts:

```bash
curl -s "http://192.168.1.2:30020/api/nginx/proxy-hosts" \
  -H "Authorization: Bearer ${TOKEN}" | jq
```

List certificates:

```bash
curl -s "http://192.168.1.2:30020/api/nginx/certificates" \
  -H "Authorization: Bearer ${TOKEN}" | jq
```

## Testing backend reachability from NPM

To debug `502 Bad Gateway`, test from inside the NPM container:

```bash
docker exec -it ix-nginx-proxy-manager-npm-1 sh
```

Then:

```sh
getent hosts jellyfin
curl -i http://jellyfin:8096
```

For Jellyfin, a good response can be:

```http
HTTP/1.1 302 Found
Server: Kestrel
Location: web/
```

If this works, NPM can reach the backend.

Do not use HTTPS to the Jellyfin backend unless Jellyfin itself is configured for HTTPS:

```sh
curl -k -i https://jellyfin:8096
```

An error like:

```text
wrong version number
```

means the backend is HTTP, not HTTPS. Use:

```yaml
npm.proxy.scheme: "http"
```

## Development

Run locally:

```bash
export NPM_BASE_URL="http://192.168.1.2:30020/api"
export NPM_EMAIL="auto.proxy@example.com"
export NPM_PASSWORD="change-me"
export LOG_LEVEL="debug"
export DOCKER_SOCKET_PATH="/var/run/docker.sock"
export LABELS_PREFIX="npm.proxy."

go run ./cmd/npm-docker-auto-proxy
```

Build binary:

```bash
CGO_ENABLED=0 GOOS=linux go build -trimpath -o npm-docker-auto-proxy ./cmd/npm-docker-auto-proxy
```

## Notes

This project intentionally avoids `switch`, `else`, and `else if` in Go code.

Preferred patterns:

```text
- early return
- map aliases
- map handlers
- small validation functions
```