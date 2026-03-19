# Deployment Guide

This project can be deployed with Docker or Podman, or as a standalone Go binary.

## 1. Deployment with Docker or Podman (Recommended)

The repository already contains:
- `compose.yaml` with `mediamtx` and `auth` services
- `mediamtx.yml` configured to call `http://auth:8080/api/auth`

### Prerequisites

- Docker with Docker Compose (`docker compose`)
- or Podman with Compose support (`podman compose`)

### Start Services

```bash
docker compose up --build -d
# or
podman compose up --build -d
```

This starts:
- MediaMTX on ports like `1935` (RTMP), `8554` (RTSP), `8888` (HLS), etc.
- auth web/API service on `127.0.0.1:8080`

### Check Status

```bash
docker compose ps
# or
podman compose ps
```

### Get Initial Admin Password

By default credentials are `admin`
But if you want you can see logs

```bash
docker compose logs auth
# or
podman compose logs auth
```

Look for a line like:
- `admin password: <value>`

### Persistent Data

- Auth DB is stored in named volume `data` mounted at `/data/auth.db`.

### Stop Services

```bash
docker compose down
# or
podman compose down
```

If you also want to remove persisted auth data:

```bash
docker compose down -v
# or
podman compose down -v
```

## 2. Standalone Binary Deployment

### Prerequisites

- Go 1.24+
- A running MediaMTX instance reachable from this service

### Build

```bash
go build -o mediamtx-auth .
```

### Run

```bash
./mediamtx-auth --db ./auth.db
```

The service listens on `:8080` by default.

### Wire MediaMTX to Auth Service

In your MediaMTX config, set:

```yaml
authMethod: http
authInternalUsers: []
authHTTPAddress: http://<auth-host>:8080/api/auth
```

## 3. Test if it works

1. Open `http://<auth-host>:8080/login`.
2. Log in with admin credentials. 
3. Create a namespace and user in `/admin`.
4. Test publishing with generated stream key (example helper script in `test_stream.sh`).

