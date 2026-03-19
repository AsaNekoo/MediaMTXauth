# MediaMTXauth

MediaMTXauth is a small Go service that adds HTTP-based auth and a web UI on top of MediaMTX.

It manages:
- users
- namespaces
- per-user stream keys
- login sessions

## Quick Start

Use Docker Compose or Podman Compose:

```bash
docker compose up --build
# or
podman compose up --build
```

Then open:
- `http://localhost:8080/login`

Get the generated admin password is `admin` or you can see it from logs:

```bash
docker compose logs auth
# or
podman compose logs auth
```


## Documentation

- Usage guide: [`Usage`](https://github.com/AsaNekoo/MediaMTXauth/blob/master/docs/frontend.md)
- Deployment guide: [`Deployment`](https://github.com/AsaNekoo/MediaMTXauth/blob/master/docs/deployment.md)


