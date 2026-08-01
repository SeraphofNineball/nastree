# NAStree

A WizTree/WinDirStat-style disk usage visualizer for a NAS, built to run in Docker
and serve a web dashboard.

- **Backend**: Go. Walks a mounted path on an interval (or on demand), stores
  results in SQLite, serves a small JSON API.
- **Frontend**: Vue 3 + ECharts. Folder table, file-type breakdown, and a
  treemap, navigable by clicking into folders.

## Run it

### Option A: build from source (dev machine)

Edit `docker-compose.yml` and point the volume at the path you want scanned:

```yaml
volumes:
  - /path/to/your/nas/share:/data:ro
  - nastree-db:/db
```

Then:

```sh
docker compose up -d --build
```

### Option B: pull the published image (NAS)

Every push to `main` builds a multi-arch (amd64 + arm64) image and publishes
it to GitHub Container Registry via `.github/workflows/docker-publish.yml`.

On the NAS, copy just `docker-compose.nas.yml` (no source checkout needed),
edit its volume path, then:

```sh
docker compose -f docker-compose.nas.yml up -d
```

To pull a newer version later: `docker compose -f docker-compose.nas.yml pull && docker compose -f docker-compose.nas.yml up -d`.

Open `http://<host>:8080` either way. The first scan starts immediately and
can take a while on a large share; the dashboard shows a "no scan yet"
message until it finishes.

## Configuration (environment variables)

| Variable | Default | Meaning |
|---|---|---|
| `SCAN_PATH` | `/data` | Path inside the container to scan (mount your NAS share here) |
| `SCAN_INTERVAL` | `6h` | How often to rescan in the background (Go duration string) |
| `RETAIN_SCANS` | `5` | How many past scans to keep in the DB (for future history/trend views) |
| `DB_PATH` | `/db/nastree.db` | SQLite database location |
| `LISTEN_ADDR` | `:8080` | HTTP listen address |

A manual rescan can also be triggered from the dashboard's "Rescan" button,
which hits `POST /api/scan/trigger`.

## Project layout

```
backend/    Go module: scanner, SQLite store, HTTP API, scan scheduler
frontend/   Vue 3 + Vite dashboard
docker/     Dockerfile (multi-stage: build frontend + backend, alpine runtime)
```

## Local development

Backend (requires Go 1.23+; not needed if you only use Docker):

```sh
cd backend
SCAN_PATH=/some/local/path go run ./cmd/server
```

Frontend (dev server proxies `/api` to `localhost:8080`, see `vite.config.ts`):

```sh
cd frontend
npm install
npm run dev
```
